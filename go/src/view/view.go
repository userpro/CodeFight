package view

func GetContent(roomtoken string) string {
    var v1 = `
    <!doctype html>
    <html lang="zh-CN">

    <head>
      <meta charset="utf-8">
      <title>WebSocket</title>
    </head>

    <body>
      <p id="output"></p>

      <script>
    `

    var v2 = `
        var loc = window.location;
        var uri = 'ws:';

        if (loc.protocol === 'https:') {
          uri = 'wss:';
        }
        uri += '//' + loc.host;
        uri += '/ws';

        ws = new WebSocket(uri)

        ws.onopen = function() {
          console.log('Connected')
          ws.send(rtk);
        }

        ws.onmessage = function(evt) {
          console.log(JSON.parse(evt.data))
          // ws.send('1') // recv message ok
        }

        console.log(rtk)
      </script>
    </body>

    </html>
    `

    return v1 + "var rtk = '" + roomtoken + "';" + v2
}