import asyncdispatch, jester, strutils, net, httpclient, os, zippy/ziparchives, random, browsers, ws, ws/jester_extra

proc getOpenPort(): Port =
  let socket = newSocket()
  socket.bindAddr(Port(0))
  let (_, port) = socket.getLocalAddr()
  socket.close()

  return port

proc generateRandomString(length: int): string =
    const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    var result = ""
    
    for _ in 0 .. length - 1:
        result.add(alphabet[rand(0 .. alphabet.len - 1)])
    
    return result

let tempDir = getTempDir()
var NEPTUNE_URL = "https://github.com/uwu/neptune/archive/refs/heads/master.zip"
var client = newHttpClient()
# We'll end up changing this conditionally based on platform.
var tidalDirectory: string
case hostOS:
    of "windows":
        tidalDirectory = joinPath(getEnv("localappdata"), "TIDAL")

        for _, path in walkDir(tidalDirectory):
          if path.startsWith("app-"):
            tidalDirectory = joinPath(tidalDirectory, path, "resources")
            break

    of "macosx":
        tidalDirectory = "/Applications/TIDAL.app/Contents/Resources"
    else:
        quit()

router myrouter:
  post "/install":
    if request.body != "": NEPTUNE_URL = request.body

    try:
      client.downloadFile(NEPTUNE_URL, joinPath(tempDir, "neptune.zip"))
      extractAll(joinPath(tempDir, "neptune.zip"), joinPath(tempDir, "neptune-unzipped"))
      removeFile(joinPath(tempDir, "./neptune.zip"))
    except:
      discard
    
    moveDir(joinPath(tempDir, "neptune-unzipped/neptune-master/injector"), tidalDirectory & "/app")
    removeDir(joinPath(tempDir, "neptune-unzipped"))
    moveFile(tidalDirectory & "/app.asar", tidalDirectory & "/original.asar")
      
    resp Http200, {"Access-Control-Allow-Origin":"*"}, "done"
  post "/uninstall":
    removeDir(tidalDirectory & "/app")
    moveFile(tidalDirectory & "/original.asar", tidalDirectory & "/app.asar")

    resp Http200, {"Access-Control-Allow-Origin":"*"}, "done"
  get "/status":
    resp Http200, {"Access-Control-Allow-Origin":"*"}, if fileExists(tidalDirectory & "/original.asar"): "installed" else: "not installed"
  get "/ws": # This route exists so that when the installer frontend closes, the server does too.
    try:
      var ws = await newWebSocket(request)
      
      while ws.readyState == Open:
        discard await ws.receiveStrPacket()
    except:
      quit()
    


proc main() =
  let port = getOpenPort()
  randomize()
  let installerKey = generateRandomString(10)
  let settings = newSettings(port=port, appName="/" & installerKey, bindAddr = "127.0.0.1")
  var jester = initJester(myrouter, settings=settings)
  openDefaultBrowser("https://neptune.uwu.network/install#" & $(port) & "/" & installerKey)
  jester.serve()

when isMainModule:
  main()