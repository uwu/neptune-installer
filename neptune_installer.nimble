# Package

version       = "0.0.1"
author        = "toonlink"
description   = "A simple installer for the TIDAL client mod neptune, leveraging the web."
license       = "Ms-PL"
srcDir        = "src"
bin           = @["neptune_installer"]


# Dependencies

requires "nim >= 1.6.14"
requires "jester == 0.6.0"
requires "zippy == 0.10.10"
requires "ws == 0.5.0"
requires "puppy >= 2.0.3"