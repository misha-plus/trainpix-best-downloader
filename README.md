TrainPix.org wallpapers
=======================

This is downloader for train pictures from https://trainpix.org/voting.php?show=results.

## Dependencies
* __Go.__ This project works with Golang v1.11.1, further Go versions probably should work.
* __waifu2x-mac.__ *[only for scaler]* waifu2x allows to denoise JPEG images and upscale them. This is waifu2x version optimized to run with Metal and macOS. On other operation system you can consider __kaifu2x__ or original __waifu2x__. https://github.com/imxieyi/waifu2x-mac
* __OpenJPEG.__ *[only for scaler]* This is JPEG2000 library, it allows to save 2x scaled image to file with size near 800kB. macOS supports JPEG2000 by default, on other operation systems you probably should consider alternative file format. http://www.openjpeg.org/.
To intstall it on Mac run `brew install openjpeg` 

## Fetcher usage
```
Usage of fetcher:
  -dir string
    	where we should save pictures (default is current working directory)
  -horiz
    	allow download horizontal pictures? [-horiz=false], [-horiz=true] (default true)
  -pages int
    	"-pages 10" will take pictures from 10 first pages, if "pages" will be -1 then crawler will take all pages (default 10)
  -upscales string
    	directory where upscaled files stored. With this option you can delete downloaded images keeping only upscales
  -vert
    	allow download vertical pictures? [-vert=false], [-vert=true]
```


Run `go run main.go -pages 10 -dir resultsPath -upscales upscalesPath` to download horizontal pictures from 10 first pages to directory `resultsPath`. Fetcher don't try to download already downloaded or upscaled pictures. `-upscales` is optional. Also, you can delete downloaded photos after upscalem, fetcher will look to upscaled pictures in this case.

## Scaler usage
You should set constants to appropriate in file `scaler/main.go` for `waifu2xmacCmd` and
`opjCompressCmd`.

`go run scaler/main.go -i pathToDownloadedPicsDir -o pathToUpscaledPicsDir`

This will take pictures from `pathToDownloadedPicsDir` directory, upscale them to 2x and convert pictures to JPEG2000 format and save it to `pathToUpscaledPicsDir` directory. If some image already upscaled then upscaler will skip it.
