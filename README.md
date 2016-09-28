# httpsocket loader

Simple load generator for httpsocket HTTP-over-websocket proxy

## Usage
```
git clone https://github.com/koroandr/httpsocket_loader.git
cd httpsocket_loader
bin/httpsocket_loader -url ws://websocket.example.com/ws -substitutions substitutions.json -proc 5 -origin=http://example.com
```

## Parameters:

 * `-data string` - Data file (default "data.log")
 * `-debug` - Show debug output
 * `-help` - Print this help
 * `-origin string` - Origin header
 * `-proc int` - Number of processes to run (default 1)
 * `-rotate` - cycle logs
 * `-sleep int` - Sleep time between requests in milliseconds (default 100)
 * `-substitutions string` - Data file
 * `-url string` - WebSocket endpoint url
