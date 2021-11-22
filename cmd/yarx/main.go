package main

import (
	"fmt"
	"github.com/kataras/golog"
	"github.com/kataras/pio"
	"github.com/remeh/sizedwaitgroup"
	"github.com/urfave/cli/v2"
	"github.com/zema1/yarx"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const helpTpl = `
_____.___.                    
\__  |   |____ __________  ___
 /   |   \__  \\_  __ \  \/  /
 \____   |/ __ \|  | \/>    < 
 / ______(____  /__|  /__/\_ \
 \/           \/            \/ - v{{.Version}}

Github: https://github.com/zema1/yarx

{{.Name}} {{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{if .VisibleFlags}}GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

// todo: status 201 || status 200 问题

func main() {
	app := &cli.App{
		Name:    "Yarx",
		Usage:   "launch a rogue server according to yaml pocs",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "pocs",
				Aliases:  []string{"p"},
				Usage:    "load pocs from this dir",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Usage:   "the http server listen address",
				Value:   "127.0.0.1:7788",
			},
			&cli.StringFlag{
				Name:    "root",
				Aliases: []string{"r"},
				Usage:   "load files form this directory if the requested path is not found",
			},
			&cli.StringFlag{
				Name:  "\r\t\t\t",
				Usage: "`\r`",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "verbose mode, which is  equivalent to --log-level debug",
				Value:   false,
			},
		},

		EnableBashCompletion:   true,
		HideHelp:               false,
		HideHelpCommand:        true,
		HideVersion:            true,
		UseShortOptionHandling: false,
	}
	cli.AppHelpTemplate = helpTpl
	app.Action = func(c *cli.Context) error {
		if c.Bool("verbose") {
			golog.SetLevel("debug")
		}
		now := time.Now()

		loader := &yarx.Yarx{}
		pocDir := c.String("pocs")
		pocPaths, err := getPocs(pocDir)
		if err != nil {
			return err
		}
		var errRules []*ErrorRule
		golog.Infof("loading pocs from %s, count: %d", c.String("pocs"), len(pocPaths))
		golog.Infof("begin to run parallel analysis")
		wg := sizedwaitgroup.New(runtime.NumCPU())
		var mu sync.Mutex
		for _, pocPath := range pocPaths {
			pocPath := pocPath
			wg.Add()
			go func() {
				defer wg.Done()
				golog.Debugf("loading %s", pocPath)
				err = loader.ParseFile(pocPath)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errRules = append(errRules, &ErrorRule{
						Filename: filepath.Base(pocPath),
						Err:      err,
					})
				}
			}()
		}
		wg.Wait()

		handler := loader.HTTPHandler()
		//handler.OnRuleMatch(func(e *yarx.ScanEvent) {
		//})
		handler.OnPocMatch(func(e *yarx.ScanEvent) {
			coloredOutput := pio.Rich(e.String(), pio.Red)
			golog.Info(coloredOutput)
		})
		if c.String("root") != "" {
			handler.SetStaticDir(c.String("root"))
		}
		fmt.Println()
		if len(errRules) != 0 {
			maxLength := getMaxLineLength(errRules) + 2
			RedPrintf("[Unsupported]\n")
			for _, eRule := range errRules {
				blank := genBlank(maxLength - len(eRule.Filename))
				errInfo := eRule.Err.Error()
				if strings.Contains(errInfo, "\n") {
					errInfo = errInfo[:strings.Index(errInfo, "\n")]
				}
				_, _ = fmt.Fprintf(os.Stderr, "  - %s%s%s\n", eRule.Filename, blank, errInfo)
			}
			fmt.Println()
		}

		listen := c.String("listen")
		lis, err := net.Listen("tcp", listen)
		if err != nil {
			return err
		}

		GreenPrintf("Analysis successfully in %s\n\n", time.Since(now).String())
		printStatus := func(k, v string) {
			fmt.Printf("%s: \t%s\n", k, pio.Rich(v, pio.Yellow, pio.Bold))
		}
		dir, err := filepath.Abs(pocDir)
		if err != nil {
			dir = pocDir
		}
		printStatus("- PocDir", dir)
		printStatus("- PocLoaded", strconv.Itoa(len(loader.Chains())))
		printStatus("- RouteCount", strconv.Itoa(len(handler.Routes())))
		printStatus("- ListenAddr", "http://"+listen)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			for range sigCh {
				fmt.Println()
				golog.Warnf("received interrupt signal, exiting...")
				_ = lis.Close()
			}
		}()
		err = http.Serve(lis, handler)
		if err != nil && err != http.ErrServerClosed && !strings.Contains(err.Error(), "use of closed network connection") {
			return err
		}
		golog.Infof("server stopped")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		RedPrintf(err.Error())
	}
}

type ErrorRule struct {
	Filename string
	Err      error
}

func getPocs(dir string) ([]string, error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, info := range infos {
		if strings.HasSuffix(info.Name(), "yml") {
			ret = append(ret, filepath.Join(dir, info.Name()))
		}
	}
	return ret, nil
}

func getMaxLineLength(rules []*ErrorRule) int {
	var maxLength = 0
	for _, r := range rules {
		if len(r.Filename) > maxLength {
			maxLength = len(r.Filename)
		}
	}
	return maxLength
}

func genBlank(i int) string {
	a := ""
	for ; i > 0; i-- {
		a += " "
	}
	return a
}

func GreenPrintf(format string, a ...interface{}) {
	target := fmt.Sprintf(format, a...)
	fmt.Print(pio.Rich(target, pio.Green))
}

func RedPrintf(format string, a ...interface{}) {
	target := fmt.Sprintf(format, a...)
	fmt.Print(pio.Rich(target, pio.Red))
}

func YellowPrintf(format string, a ...interface{}) {
	target := fmt.Sprintf(format, a...)
	fmt.Print(pio.Rich(target, pio.Yellow))
}
