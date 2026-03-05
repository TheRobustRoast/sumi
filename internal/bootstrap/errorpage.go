package bootstrap

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"strings"
)

// ServeErrorPage starts an HTTP server on port 7777 that shows the install log.
// It blocks until the server is stopped. Call this in a goroutine.
func ServeErrorPage(logPath string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log, err := os.ReadFile(logPath)
		if err != nil {
			log = []byte(fmt.Sprintf("Log file not available (%v)", err))
		}

		lines := strings.Split(string(log), "\n")
		var tail string
		if len(lines) > 80 {
			tail = strings.Join(lines[len(lines)-80:], "\n")
		} else {
			tail = string(log)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, errorPageTemplate, html.EscapeString(logPath), html.EscapeString(tail), html.EscapeString(string(log)))
	})

	_ = http.ListenAndServe(":7777", nil)
}

// LocalIP returns the machine's first non-loopback IPv4 address.
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "localhost"
}

const errorPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>sumi :: install error</title>
  <style>
    *{box-sizing:border-box}
    body{background:#0a0a0a;color:#d4d4d4;font-family:monospace;padding:2em;max-width:1400px;margin:0 auto}
    h1{color:#f7768e;margin:0 0 .4em}
    .meta{color:#6a6a6a;font-size:.85em;margin-bottom:1.5em}
    .meta span{color:#8a8a8a}
    .tabs{display:flex;gap:.5em;margin-bottom:1em}
    .tabs button{background:#1a1a1a;border:1px solid #2a2a2a;color:#d4d4d4;
                  padding:.4em 1.2em;border-radius:6px;cursor:pointer;font:inherit;font-size:.9em}
    .tabs button.active{background:#2a2a2a;color:#f7768e;border-color:#f7768e}
    pre{background:#1a1a1a;padding:1.5em;border-radius:8px;overflow:auto;
         line-height:1.6;font-size:.8em;white-space:pre-wrap;word-break:break-all;
         border:1px solid #2a2a2a;margin:0}
  </style>
</head>
<body>
  <h1>sumi :: installation failed</h1>
  <p class="meta">log <span>%s</span></p>
  <div class="tabs">
    <button class="active" onclick="show('tail')" id="tab-tail">Error (last 80 lines)</button>
    <button onclick="show('full')" id="tab-full">Full Log</button>
  </div>
  <pre id="pane-tail">%s</pre>
  <pre id="pane-full" style="display:none">%s</pre>
  <script>
  function show(id){
    ['tail','full'].forEach(function(p){
      document.getElementById('pane-'+p).style.display=p===id?'':'none';
      document.getElementById('tab-'+p).className=p===id?'active':'';
    });
  }
  </script>
</body>
</html>`
