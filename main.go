// md3dump - examine quake 3 md3 file
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
)

const (
	MAX_QPATH     = 64
	MD3_XYZ_SCALE = 1 / 64.0
)

type Header struct {
	Magic     int32
	Version   int32
	Name      [MAX_QPATH]uint8
	Flags     int32
	NumFrames int32
	NumTags   int32
	NumSurfs  int32
	NumSkins  int32
	OfsFrames int32
	OfsTags   int32
	OfsSurfs  int32
	OfsEOF    int32
}

type Frame struct {
	MinBounds   [3]float32
	MaxBounds   [3]float32
	LocalOrigin [3]float32
	Radius      float32
	Name        [16]uint8
}

type Tag struct {
	Name   [MAX_QPATH]uint8
	Origin [3]float32
	Axis   [3][3]float32
}

type Surf struct {
	Magic        int32
	Name         [MAX_QPATH]uint8
	Flags        int32
	NumFrames    int32
	NumShaders   int32
	NumVerts     int32
	NumTris      int32
	OfsTris      int32
	OfsShaders   int32
	OfsST        int32
	OfsXYZNormal int32
	OfsEnd       int32
}

type XYZNormal struct {
	XYZ    [3]int16
	Normal int16
}

type Shader struct {
	Name  [MAX_QPATH]uint8
	Index int32
}

func xyz(c [3]int16) [3]float32 {
	var v [3]float32
	v[0] = float32(c[0]) * MD3_XYZ_SCALE
	v[1] = float32(c[1]) * MD3_XYZ_SCALE
	v[2] = float32(c[2]) * MD3_XYZ_SCALE
	return v
}

func normal(n int16) [3]float32 {
	lat := float64((n>>8)&0xFF) * (2 * math.Pi) / 0xFF
	lng := float64((n>>0)&0xFF) * (2 * math.Pi) / 0xFF
	var v [3]float32
	v[0] = float32(math.Cos(lat) * math.Sin(lng))
	v[1] = float32(math.Sin(lat) * math.Sin(lng))
	v[2] = float32(math.Cos(lng))
	return v
}

func magic(m int32) string {
	b := make([]byte, 4)
	b[0] = byte(m)
	b[1] = byte(m >> 8)
	b[2] = byte(m >> 16)
	b[3] = byte(m >> 24)
	return string(b)
}

func cstr(b []byte) string {
	nul := bytes.Index(b, []byte{0})
	return string(b[:nul])
}

func dump(b []byte) {
	p := fmt.Println
	pf := fmt.Printf

	//
	// Header
	//
	var m Header
	r := bytes.NewReader(b)
	err := binary.Read(r, binary.LittleEndian, &m)
	if err != nil {
		log.Fatal(err)
	}

	pf("Magic = '%s'\n", magic(m.Magic))
	p("Version =", m.Version)
	pf("Name = '%s'\n", cstr(m.Name[:]))
	p("Flags =", m.Flags)
	p("NumFrames =", m.NumFrames)
	p("NumTags =", m.NumTags)
	p("NumSurfs =", m.NumSurfs)
	p("NumSkins =", m.NumSkins)
	pf("OfsFrames = %#x\n", m.OfsFrames)
	pf("OfsTags = %#x\n", m.OfsTags)
	pf("OfsSurfs = %#x\n", m.OfsSurfs)
	pf("OfsEOF = %#x\n", m.OfsEOF)

	//
	// Frames
	//
	f := make([]Frame, m.NumFrames)
	r = bytes.NewReader(b[m.OfsFrames:])
	err = binary.Read(r, binary.LittleEndian, &f)
	if err != nil {
		log.Fatal(err)
	}

	for i, fr := range f {
		p("Frame", i)
		p(" MinBounds =", fr.MinBounds)
		p(" MaxBounds =", fr.MaxBounds)
		p(" LocalOrigin =", fr.LocalOrigin)
		p(" Radius =", fr.Radius)
		p(" Name =", cstr(fr.Name[:]))
	}

	//
	// Surfaces
	//
	surfs := make([]Surf, m.NumSurfs)
	r = bytes.NewReader(b[m.OfsSurfs:])
	err = binary.Read(r, binary.LittleEndian, &surfs)
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range surfs {
		p("Surface", i)
		pf(" Magic = '%s'\n", magic(s.Magic))
		pf(" Name = '%s'\n", cstr(s.Name[:]))
		p(" Flags =", s.Flags)
		p(" NumFrames =", s.NumFrames)
		p(" NumShaders =", s.NumShaders)
		p(" NumVerts =", s.NumVerts)
		p(" NumTris =", s.NumTris)
		pf(" OfsTris = %#x\n", s.OfsTris)
		pf(" OfsShaders = %#x\n", s.OfsShaders)
		pf(" OfsST = %#x\n", s.OfsST)
		pf(" OfsXYZNormal = %#x\n", s.OfsXYZNormal)
		pf(" OfsEnd = %#x\n", s.OfsEnd)

		//
		// Shader
		//
		//shaders := make([]Shader, s.NumShaders)
		//r = bytes.NewReader(b)
		//r.Next(s.OfsShader)

		//
		// XYZNormal
		//
		verts := make([]XYZNormal, s.NumVerts)
		r = bytes.NewReader(b[s.OfsXYZNormal:])
		err = binary.Read(r, binary.LittleEndian, &verts)
		if err != nil {
			log.Fatal(err)
		}

		for _, v := range verts {
			p(" Vertex =", xyz(v.XYZ))
		}
		for _, v := range verts {
			p(" Normal =", normal(v.Normal))
		}
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	var f *os.File
	if flag.NArg() > 0 {
		var err error
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		f = os.Stdin
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	dump(b)
}
