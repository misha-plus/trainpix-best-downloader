package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lmittmann/ppm"
)

const waifu2xmacCmd = "waifu2xmetal"
const cjpegCmd = "/usr/local/opt/mozjpeg/bin/cjpeg"
const opjCompressCmd = "opj_compress"

func isFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func main() {
	in := flag.String("i", "", "input directory")
	out := flag.String("o", "", "output directory")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.Printf("Using %s for input, %s for output", *in, *out)
	inDir, err := os.Open(*in)
	if err != nil {
		log.Fatal("Unable to open input directory", err)
	}
	defer inDir.Close()
	files, err := inDir.Readdir(0)
	if err != nil {
		log.Fatal("Unable to read files", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".jpg") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(file.Name()), ".tmp.jpg") {
			continue
		}

		outjpg2000path := filepath.Join(*out, file.Name()) + ".15opj.jp2"
		if ok, _ := isFileExist(outjpg2000path); ok {
			log.Printf("File %s already upscaled", file.Name())
			continue
		}

		err := transformFile(filepath.Join(*in, file.Name()), outjpg2000path)
		if err != nil {
			log.Printf("[Warning] Unable to transform: %v", err)
			return
		}
	}
}

func transformFile(infilepath, outfilepath string) error {
	tmpPNGFile, err := ioutil.TempFile("", "*.png")
	if err != nil {
		return err
	}
	tmpPNGFile.Close()
	defer os.Remove(tmpPNGFile.Name())

	tmpPPMFile, err := ioutil.TempFile("", "*.ppm")
	if err != nil {
		return err
	}
	tmpPPMFile.Close()
	defer os.Remove(tmpPPMFile.Name())

	log.Printf("Upscaling file %s", infilepath)

	err = upscaleJPGToPNG(infilepath, tmpPNGFile.Name())
	if err != nil {
		return fmt.Errorf("unable to upscale: %v", err)
	}

	// err = convertPNGToJPEGMoz95(outpngpath, outjpgmozpath)
	// if err != nil {
	// 	log.Printf("Unable to convert PNG to JPEG (Moz): %v", err)
	// 	continue
	// }

	// err = convertPNGToJPEGGo95(outpngpath, outjpggopath)
	// if err != nil {
	// 	log.Printf("Unable to convert PNG to JPEG (Go): %v", err)
	// 	continue
	// }

	err = convertPNGToPPMGo(tmpPNGFile.Name(), tmpPPMFile.Name())
	if err != nil {
		return fmt.Errorf("unable to convert PNG to PPM (Go): %v", err)
	}

	tmpJP2Filepath := outfilepath + ".tmp.jp2"
	err = os.Remove(tmpJP2Filepath)
	if err != nil && os.IsExist(err) {
		return err
	}
	defer os.Remove(tmpJP2Filepath)

	err = convertPPMToJPEG2000OPJ15(tmpPPMFile.Name(), tmpJP2Filepath)
	if err != nil {
		return fmt.Errorf("Unable to convert PPM to JPEG2000 (OPJ): %v", err)
	}
	return os.Rename(tmpJP2Filepath, outfilepath)
}

func upscaleJPGToPNG(inpath, outpath string) error {
	cmd := exec.Command(
		waifu2xmacCmd,
		"-t", "p",
		"-s", "2",
		"-n", "4",
		"-i", inpath,
		"-o", outpath,
	)
	return cmd.Run()
}

func convertPNGToJPEGMoz95(in, out string) error {
	cmd := exec.Command(
		cjpegCmd,
		"-optimize",
		"-progressive",
		"-quality", "95",
		"-outfile", out,
		in,
	)
	return cmd.Run()
}

func convertPNGToJPEGGo95(inpath, outpath string) error {
	in, err := os.Open(inpath)
	if err != nil {
		return err
	}
	defer in.Close()
	img, err := png.Decode(in)
	if err != nil {
		return err
	}
	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer out.Close()
	return jpeg.Encode(out, img, &jpeg.Options{95})
}

func convertPNGToPPMGo(inpath, outpath string) error {
	in, err := os.Open(inpath)
	if err != nil {
		return err
	}
	defer in.Close()
	img, err := png.Decode(in)
	if err != nil {
		return err
	}
	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer out.Close()
	return ppm.Encode(out, img)
}

func convertPPMToJPEG2000OPJ15(inpath, outpath string) error {
	cmd := exec.Command(
		opjCompressCmd,
		"-r", "15",
		"-i", inpath,
		"-o", outpath,
	)
	return cmd.Run()
}
