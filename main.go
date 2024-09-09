/* Chiwigames Community games
Copyright (C) 2024 Anïs Khoury Ribas

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type user struct {
	name      string
	connected bool
	contador  int
	m         sync.Mutex
}
type channel struct {
	name  string
	users []user
}

// var users [string]websocket.Conn
var users = make(map[net.Conn]user)
var channels = make(map[string]channel)
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func chat(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	//fmt.Printf("%+v\n", users)
	//fmt.Println(reflect.TypeOf(c))
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	//Try login with user & password
	var userClient string
	var passClient string

	mt, getUserPassword, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		println("%s", err.Error())
		return
	} else {
		var result = strings.Split(string(getUserPassword), " ")
		userClient = result[0]
		passClient = result[1]
		//if there is no username or password, then stop
		if userClient == "" || passClient == "" {
			return
		}
		println(mt, "user:", userClient, " pass:", passClient)
	}
	users[c.NetConn()] = user{userClient, true, 0, sync.Mutex{}}

	for {
		mt, message, err := c.ReadMessage()

		if entry, ok := users[c.NetConn()]; ok {
			var n = users[c.NetConn()].m
			n.Lock()
			// Then we modify the copy
			entry.contador = entry.contador + 1
			// Then we reassign map entry
			users[c.NetConn()] = entry
			n.Unlock()
		}

		if err != nil {
			log.Println("read:", err)
			println("%s", err.Error())
			break
		}
		fmt.Printf("%+v\n", users)

		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
	fmt.Println("Desconexio del socket")
	//TODO: Delete conection from users
	//TODO: Delete conection from channels
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/chat")
}

func main() {
	showtext("prova")
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/chat", chat)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var userClient = document.getElementById("userClient");
    var passwordClient = document.getElementById("passClient");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
            var resultsend =""+userClient.value+" "+passClient.value;
            ws.send(resultsend);
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
User<input id="userClient" type="text" value="user1">
Pass<input id="passClient" type="text" value="pass">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
