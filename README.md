# httpsocket loader

Simple load generator for httpsocket HTTP-over-websocket proxy

## Usage
```
git clone https://github.com/koroandr/httpsocket_loader.git
cd httpsocket_loader
bin/httpsocket_loader -url ws://websocket.example.com/ws -substitutions substitutions.json -proc 5 -origin=http://example.com
```

## Command-line parameters:

 * `-data string` - Data file (default "data.log")
 * `-debug` - Show debug output
 * `-help` - Print this help
 * `-origin string` - Origin header
 * `-proc int` - Number of processes to run (default 1)
 * `-rotate` - cycle logs
 * `-sleep int` - Sleep time between requests in milliseconds (default 100)
 * `-substitutions string` - Data file
 * `-url string` - WebSocket endpoint url

## Example data file
```
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params"}
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params with substitution {{demo}}"}
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params with per-process substitution {{proc_demo}}"}
```

## Example substitutions file
```
{
  "demo": "demo substitution",
  "proc": ["proc0", "proc1", "proc2"]
}
```

Substitution file is a plain JSON object with string keys and string or array of strings values. Substitution keys are placed in data files within double curly brackets.

* If the value is a string, then `key->value` substitution is applied
* If the value is an array of strings, then `key->value[proc_num % value.length]` substitution is applied, where proc_num is a number of current loader thread

The above data file with above substitutions for process with number 0 will turn into

```
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params"}
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params with substitution demo substitution"}
{"jsonrpc":"2.0","id":"14748814020958653","method":"someMethod","params":"some params with per-process substitution proc0"}
```
