package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/bypass"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/defaults"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
)

func screenshot(page *rod.Page, e *rod.Element, path string) {
	a, _ := e.Eval(`
	()=>{
		x=this.getBoundingClientRect()
		console.log(this,  x)
		return {bottom: x.bottom, height: x.height, left:x.left, right:x.right, top:x.top, width:x.width, x:x.x, y:x.y}
	}
	`)

	opts := &proto.PageCaptureScreenshot{
		Format: "png",
		Clip: &proto.PageViewport{
			X:      a.Value.Get("x").Num(),
			Y:      a.Value.Get("y").Num(),
			Width:  a.Value.Get("width").Num(),
			Height: a.Value.Get("height").Num(),
			Scale:  1,
		},
		FromSurface: true,
	}
	bin, _ := page.Screenshot(true, opts)
	utils.OutputFile(path, bin)
}
func results(page *rod.Page, basePath string) {
	time.Sleep(1 * time.Second)
	results := page.MustElements("h2 > a")
	for i, r := range results {
		t := r.MustText()
		fmt.Println(i, t)
		x := r.MustParents("li")[0]
		pathPrefix := basePath + strconv.Itoa(i) + "-" + t
		ioutil.WriteFile(pathPrefix+".link", []byte(*r.MustAttribute("href")), 0777)
		ioutil.WriteFile(pathPrefix+".html", []byte(x.MustHTML()), 0777)
		ioutil.WriteFile(pathPrefix+".txt", []byte(x.MustText()), 0777)
		screenshot(page, x, pathPrefix+".png")
	}
}
func run(search string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r, search)
		}
	}()

	device := devices.IPadPro
	browser := rod.New().DefaultDevice(device).Timeout(10 * time.Minute).Trace(true).MustConnect()
	defer browser.MustClose()

	// You can also use bypass.JS directly without rod
	fmt.Printf("js file size: %d\n\n", len(bypass.JS))

	// search := "lose belly fat"
	page := bypass.MustPage(browser)

	page.MustWaitNavigation()
	page.MustNavigate("https://bing.com/?count=100")
	time.Sleep(3 * time.Second)
	page.MustElement(`#sb_form_q`).MustWaitVisible().MustInput(search).MustPress(input.Enter)

	basePath := "results/" + search + "/" + time.Now().Format("2006-01-02T15:04:05") + "/"
	os.MkdirAll(basePath, 0777)

	time.Sleep(3 * time.Second)
	page.MustScreenshotFullPage(basePath + search + "-page1.png")
	results(page, basePath+"page1-")

	page.MustSearch(`[aria-label="Page 2"]`)
	time.Sleep(3 * time.Second)
	page.MustScreenshotFullPage(basePath + search + "-page2.png")
	results(page, basePath+"page2-")

	page.MustSearch(`[aria-label="Page 3"]`)
	time.Sleep(3 * time.Second)
	page.MustScreenshotFullPage(basePath + search + "-page3.png")
	results(page, basePath+"page3-")

	time.Sleep(5 * time.Second)
}

func main() {
	defaults.Show = false
	defaults.Slow = 1200 * time.Millisecond
	defaults.Devtools = true
	// defaults.Dir = "./data"
	b, err := ioutil.ReadFile("./Lose Belly Fat.txt")
	if err != nil {
		fmt.Println("Error reading file", err)
		return
	}
	terms := strings.Split(string(b), "\n")
	for i, search := range terms {
		fmt.Print(i, search)
		run(search)
	}

}

func printReport(page *rod.Page) {
	el := page.MustElement("#broken-image-dimensions.passed")
	for _, row := range el.MustParents("table").First().MustElements("tr:nth-child(n+2)") {
		cells := row.MustElements("td")
		key := cells[0].MustProperty("textContent")
		if strings.HasPrefix(key.String(), "User Agent") {
			fmt.Printf("\t\t%s: %t\n\n", key, !strings.Contains(cells[1].MustProperty("textContent").String(), "HeadlessChrome/"))
		} else {
			fmt.Printf("\t\t%s: %s\n\n", key, cells[1].MustProperty("textContent"))
		}
	}

	page.MustScreenshot("")
}
