package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options
var homeTemplate = template.New("")
var m = map[string]string{}

type OasisFeed struct {
	OasisId string `json:"oasis_id"`
	Message string `json:"message"`
}

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
		var oasis_feed OasisFeed
		err = json.Unmarshal(message, &oasis_feed)
		if err != nil {
			log.Println("error while unmarshaling oasis feed")
			return
		}
		m[oasis_feed.OasisId] = oasis_feed.Message
	}
}

func feed(w http.ResponseWriter, r *http.Request) {
	oasis_id := r.URL.Query().Get("oasis_id")
	if oasis_id != "" {
		feed_value, ok := m[oasis_id]
		if ok {
			delete(m, oasis_id)
			fmt.Fprint(w, feed_value)
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	address := os.Args[1]
	var addr = flag.String("addr", address, "http service address")
	flag.Parse()
	log.SetFlags(0)
	prepareTemplate(address)
	http.HandleFunc("/oasis", oasis)
	http.HandleFunc("/", home)
	http.HandleFunc("/feed", feed)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func prepareTemplate(address string) {
	var templateContent = `
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
    const oasisIdInput = document.getElementById("oasis_id");
    const taLineHeight = 20;
    stopOasisButton.disabled = true;
    var controller = null;
    var is_feed_disable = false;

    function sleep(ms) {
  		return new Promise(resolve => setTimeout(resolve, ms));
	}

    function start_poll_oasis(ms, oasis_id) {
        var poll = (promiseFn, duration) => promiseFn().then(sleep(duration).then(() => {
                if (is_feed_disable === false) {
                    poll(promiseFn, duration);
                }
            }    
        ));
        poll(() => new Promise(() => get_feed(oasis_id)), ms);
    } 

    function get_feed(oasis_id) {
        if (is_feed_disable === false) {
                const xhr = new XMLHttpRequest();
                const url = "http://%s/feed?oasis_id=" + oasis_id;
                xhr.open("GET", url);
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
        const oasis_id = oasisIdInput.value;
        start_poll_oasis(1000, oasis_id);

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
    <div id="oasis_header">Oasis id</div>
    <input type="text" id="oasis_id" name="oasis_id" minlength="4" maxlength="8" size="10" />
    <br /><br />
    <button id="start_oasis" class="button-oasis" role="button">Start</button>
    <button id="stop_oasis" class="button-oasis" role="button">Stop</button>
</body>  
</html>
`

	var formattedContent = fmt.Sprintf(templateContent, address)
	template.Must(homeTemplate.Parse(formattedContent))
}
