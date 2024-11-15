package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options
var m = map[string]string{}

func oasis(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			//log.Println("read:", err)
			break
		}
		//log.Printf("recv: %s", message)
		m["feed"] = string(message)
	}
}

func feed(w http.ResponseWriter, r *http.Request) {
	feed_value, ok := m["feed"]
	if ok {
		delete(m, "feed")
		fmt.Fprint(w, feed_value)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/oasis", oasis)
	http.HandleFunc("/", home)
	http.HandleFunc("/feed", feed)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>WS Oasis dashboard</title>

<style>
    #ws_oasis {
        resize: none;
    }

.button-oasis {
  background-color: #e1ecf4;
  border-radius: 3px;
  border: 1px solid #7aa7c7;
  box-shadow: rgba(255, 255, 255, .7) 0 1px 0 0 inset;
  box-sizing: border-box;
  color: #39739d;
  cursor: pointer;
  display: inline-block;
  font-family: -apple-system,system-ui,"Segoe UI","Liberation Sans",sans-serif;
  font-size: 13px;
  font-weight: 400;
  line-height: 1.15385;
  margin: 0;
  outline: none;
  padding: 8px .8em;
  position: relative;
  text-align: center;
  text-decoration: none;
  user-select: none;
  -webkit-user-select: none;
  touch-action: manipulation;
  vertical-align: baseline;
  white-space: nowrap;
}

.button-oasis:hover,
.button-oasis:focus {
  background-color: #b3d3ea;
  color: #2c5777;
}

.button-oasis:focus {
  box-shadow: 0 0 0 4px rgba(0, 149, 255, .15);
}

.button-oasis:active {
  background-color: #a0c7e4;
  box-shadow: none;
  color: #2c5777;
}

.button-oasis:disabled,
.button-oasis[disabled]{
  border: 1px solid #999999;
  background-color: #cccccc;
  color: #666666;
}
</style>

<script>  
window.addEventListener("load", function(evt) {
    const oasisEl = document.getElementById("ws_oasis");
    const startOasisButton = document.getElementById("start_oasis");
    const stopOasisButton = document.getElementById("stop_oasis");
    const taLineHeight = 20;
    stopOasisButton.disabled = true;
    var controller = null;
    var is_feed_disable = false;

    function sleep(ms) {
  		return new Promise(resolve => setTimeout(resolve, ms));
	}

    function start_poll_oasis(ms) {
        var poll = (promiseFn, duration) => promiseFn().then(sleep(duration).then(() => {
                if (is_feed_disable === false) {
                    poll(promiseFn, duration);
                }
            }    
        ));
        poll(() => new Promise(() => get_feed()), ms);
    } 

    function get_feed() {
        if (is_feed_disable === false) {
                const xhr = new XMLHttpRequest();
                xhr.open("GET", "http://127.0.0.1:8080/feed");
                xhr.send();
                xhr.responseType = "plain/text";
                xhr.onload = () => {
                    if (xhr.readyState == 4 && xhr.status == 200) {
                        var response = xhr.response;
                        if (response !== "") {
                            var taHeight = oasisEl.scrollHeight;
                            var numberOfLines = Math.floor(taHeight/taLineHeight);
                            if (numberOfLines > 15) {
                                oasisEl.textContent = "clearing\r\n";
                                var text = oasisEl.textContent;
                                oasisEl.textContent = text + response + '\r\n';
                            } else {
                                var text = oasisEl.textContent;
                                oasisEl.textContent = text + response + '\r\n';
                            }
                        }
                    }
                }
        }
    };

    stopOasisButton.onclick = function(evt) {
        if (controller !== null && is_feed_disable === false) {
            controller.abort('stop');
            stopOasisButton.disabled = true; 
            startOasisButton.disabled = false; 
        }
    };

    startOasisButton.onclick = function(evt) {
        controller = new AbortController();
        const abortListener = ({target}) => {
            controller.signal.removeEventListener('abort', abortListener);
            is_feed_disable = true;    
        } 
        controller.signal.addEventListener('abort', abortListener);

        is_feed_disable = false;
        start_poll_oasis(1000);

        stopOasisButton.disabled = false; 
        startOasisButton.disabled = true; 
    };

});
</script>
</head>
<body>
    <div id="oasis_header">Oasis output</div>
    <textarea id="ws_oasis" name="oasis" rows="20" cols="100" resize="none"></textarea>
    <br /><br />
    <button id="start_oasis" class="button-oasis" role="button">Start</button>
    <button id="stop_oasis" class="button-oasis" role="button">Stop</button>
</body>  
</html>
`))
