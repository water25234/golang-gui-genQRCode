//go:generate go run -tags generate gen.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/skip2/go-qrcode"
	"github.com/zserge/lorca"
)

func main() {
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 960, 720, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	// A simple way to know when UI is ready (uses body.onload event in JS)
	ui.Bind("start", func() {
		log.Println("UI is ready")
	})

	// Create and bind Go object to the UI
	g := &generator{}
	ui.Bind("generatorQRCode", g.generatorQRCode)

	// Load HTML.
	// You may also use `data:text/html,<base64>` approach to load initial HTML,
	// e.g: ui.Load("data:text/html," + url.PathEscape(html))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))

	// _, b, _, _ := runtime.Caller(0)
	// basepath := filepath.Dir(b)
	basePath, _ := os.UserHomeDir()
	os.Chdir(basePath)

	pwd, _ := os.Getwd()

	// You may use console.log to debug your JS code, it will be printed via
	// log.Println(). Also exceptions are printed in a similar manner.
	ui.Eval(`
		console.log("Hello, world!");
		console.log('Multiple values:', [1, false, {"x":5}]);
		console.log("` + pwd + `");`)
	// console.log("` + pwd + `");`

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}

// Go types that are bound to the UI must be thread-safe, because each binding
// is executed in its own goroutine. In this simple case we may use atomic
// operations, but for more complex cases one should use proper synchronization.
type generator struct {
	sync.Mutex
	pinCodeList string
	folder      string
	fileExt     string
}

type errLog struct {
	errGenQRCode []string
}

type jobChannel struct {
	index       int
	fileContent string
}

func (g *generator) generatorQRCode(pinCodeList string, folder string, fileExt string) (result string, err error) {
	g.Lock()
	defer g.Unlock()
	if len(pinCodeList) == 0 {
		return "", fmt.Errorf("flags readfile is empty")
	}

	if len(folder) == 0 {
		return "", fmt.Errorf("flags folder is empty")
	}

	if len(fileExt) == 0 {
		return "", fmt.Errorf("flags fileExt is empty")
	}

	g.pinCodeList = pinCodeList
	g.folder = folder
	g.fileExt = fileExt

	result = g.processQRCode()
	return result, nil
}

func (g *generator) processQRCode() (result string) {
	fmt.Println("--------------- start work ---------------")

	fileContentArr := strings.Split(g.pinCodeList, "\n")
	fileContentCount := len(fileContentArr)
	errGenQRCode := &errLog{}

	os.MkdirAll(g.folder, os.ModePerm)

	// channel for job
	jobChans := make(chan jobChannel, fileContentCount)

	// start workers
	wg := &sync.WaitGroup{}
	wg.Add(fileContentCount)

	// start workers
	for i := 1; i <= runtime.NumCPU(); i++ {
		go func(i int) {
			for job := range jobChans {
				g.work(job.fileContent, errGenQRCode)
				wg.Done()
			}
		}(i)
	}

	// collect job
	for i := 0; i < fileContentCount; i++ {
		jobChans <- jobChannel{
			index:       i,
			fileContent: fileContentArr[i],
		}
	}

	close(jobChans)

	wg.Wait()

	if len(errGenQRCode.errGenQRCode) > 0 {
		fmt.Println("error gen qr code failure list : ", errGenQRCode.errGenQRCode)
	}

	fmt.Println("--------------- finish work ---------------")
	return fmt.Sprintf("執行完成，請找資料夾 『 %s 』 並且確認檔案數量與內容", g.folder)
}

func (g *generator) fileSize(pingCode string) (size int64, err error) {
	fi, err := os.Stat(pingCode)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	return fi.Size(), nil
}

func (g *generator) pinCodeInfo(valueArr []string) (valueName string, valuePinCode string, err error) {
	if len(valueArr) == 1 {
		valueName = valueArr[0]
		valuePinCode = valueArr[0]
	} else if len(valueArr) == 2 {
		valueName = valueArr[0]
		valuePinCode = valueArr[1]
	} else {
		fmt.Println("value format is error")
		return "", "", fmt.Errorf("value format is error")
	}
	return valueName, valuePinCode, nil
}

func (g *generator) work(fileContent string, errGenQRCode *errLog) {

	if len(fileContent) == 0 {
		return
	}

	valueArr := strings.Split(strings.TrimSpace(fileContent), " ")
	valueName, valuePinCode, err := g.pinCodeInfo(valueArr)
	if err != nil {
		return
	}

	pingCode := g.folder + "/" + valueName + g.fileExt

	err = qrcode.WriteFile(valuePinCode, qrcode.Medium, 256, pingCode)

	if err != nil {
		fmt.Println("gen QR Code failure", pingCode)
		errGenQRCode.errGenQRCode = append(errGenQRCode.errGenQRCode, pingCode)
		return
	}

	size, err := g.fileSize(pingCode)
	if err != nil {
		fmt.Println("get file size failure", pingCode)
		errGenQRCode.errGenQRCode = append(errGenQRCode.errGenQRCode, pingCode)
		return
	}

	fmt.Println(fmt.Sprintf("file: %s, file size: %d", pingCode, size))
	return
}
