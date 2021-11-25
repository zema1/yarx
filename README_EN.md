<div align="center">
<img src="assets/images/logo.png" alt="Logo" height="140">
</div>

<p align="center">
    <a href="https://yarx.koalr.me/"><b>Demo</b></a>&nbsp;&nbsp;&nbsp;
    <a href="https://yarx.koalr.me/report.html"><b>Report</b></a> 
</p>


Yarx comes from the reverse spelling of `x-r-a-y`, and it can fully automatically generate a Server that satisfies the rules according to xray's yaml poc rules. Scanning the server with xray will get a dozen of  corresponding vulnerabilities.

![yarx-core](assets/images/core.svg)

## Feature

+ Support response mutation for status, header, body, etc.
+ Support for various matching patterns such as `=`, `contains`, `submatch`, etc.
+ Support for rendering and capturing dynamic variables and variable tracking with multi-level rules
+ Support for parsing and calling most of the defined functions
+ Reduce route conflicts with route merging and smart sorting strategies
+ Support for capturing scan events for further analysis and linkage
+ Support concurrent scans

## Try with xray

```bash
./xray webscan --plugins phantasm --html-output yarx.html --url https://yarx.koalr.me
```

![running](./assets/images/scan.gif)

After a few second, you will get a vulnerablity report  like that:  [report.html](https://yarx.koalr.me/report.html)


## Installation
+ Github Release
  
  [https://github.com/zema1/yarx/releases](https://github.com/zema1/yarx/releases) 
  Download the release suitable for your platform and run it in cli.
  
+ Compile Source
  ```bash
  git clone https://github.com/zema1/yarx
  cd yarx
  go build -o yarx ./cmd/yarx
  ```

## Usage

```bash
USAGE:
   yarx [global options] [arguments...]

GLOBAL OPTIONS:
   --pocs value, -p value    load pocs from this dir
   --listen value, -l value  the http server listen address (default: "127.0.0.1:7788")
   --root value, -r value    load files form this directory if the requested path is not found

   --verbose, -V             verbose mode, which is  equivalent to --log-level debug (default: false)
   --help, -h                show help (default: false)
```

Example：

```bash
# Create an http server on port 8080 to simulate all vulnerabilities in the pocs folder
./yarx -p ./pocs -l 0.0.0.0:8080

# Same as above but use the file in the `./www/html` folder when the request path doesn't match any poc
./yarx -p ./pocs -l 0.0.0.0:8080 -r ./www/html
```
![running](assets/images/running.png)

You can use the [pocs](./pocs) folder of this repository, or use the [https://github.com/chaitin/xray/tree/master/pocs](https://github.com/chaitin/xray/tree/master/pocs) folder of the official xray repository directly. This repository simply removes the temporarily unsupported pocs, which make no difference with the official repo except that they may print a little error message at runtime, and I will periodically sync the data to add more verified pocs.

Of course, you can load your own pocs.

## Development

Yarx can also be used as a go package

```go
yr := &yarx.Yarx{}
// err := yr.Parse([]byte("poc-data"))
err := yr.ParseFile("/path/to/a/yaml/poc")
if err != nil {
    panic(err)
}

// Each successfully loaded poc corresponds to a MutationChain
// The rule in a poc corresponds to a MutationRule
chains := yr.Chains()
rules := yr.Rules()
...

// Generate the http handler for the above rule with one click
handler := yr.HTTPHandler()

// event handler
handler.OnRuleMatch(func(e *yarx.ScanEvent) {
})
handler.OnPocMatch(func(e *yarx.ScanEvent) {
    fmt.Println(e.RemoteAddr)
    fmt.Println(e.Request)
    fmt.Println(e.Response)
    fmt.Println(e.PocMatched)
    fmt.Println(e.RuleMatched)
})

// launch the http server
http.ListenAndServe(handler, "127.0.0.1:7788")
```

## Errors Explanation

Yarx may encounter errors when parsing pocs, those pocs will not be loaded into the final http service, do not worry about that and basically the errors are these types of problems.

+ paths that are too flexible

  Paths like `{{name}}.php` and `/`, which are not distinguishable from other similar rules when used as routes (trust me, Yarx has done its best to avoid conflicting paths)

+ Does not support the cases where there are complex transformations in the ``set`` definition, e.g.

  ```yaml
  set:
    r0: randLowercase(8)
    r1: base64(r0) # Tracking this variable is too complex
  ```
  
+ Does not support the use of reverse service, i.e. yaml with `newReverse()` calls, but i plan to support it later

If you encounter other types of errors, you can submit an issue with the yaml poc and the details of the error, and I will deal with it as soon as possible.

## Roadmap

- [ ] Support for Docker deployments
- [ ] Support for POCs that rely on `newReverse` variable
- [ ] Support for POCs that rely on `request` variable
