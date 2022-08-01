package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	perlin "github.com/aquilax/go-perlin"
)

const width int = 2000
const height int = 800

type Vec3 struct {
	x, y, z float64
}

func (p *Vec3) len() float64 {
	return math.Sqrt(p.x*p.x + p.y*p.y + p.z*p.z)
}

func (p *Vec3) normalize() {
	l := p.len()
	p.x /= l
	p.y /= l
	p.z /= l
}

func (p *Vec3) dot(v Vec3) float64 {
	return p.x*v.x + p.y*v.y + p.z*v.z
}

func (p *Vec3) multiply(t float64) Vec3 {
	return Vec3{p.x * t, p.y * t, p.z * t}
}

func Cross(v1, v2 Vec3) Vec3 {
	return Vec3{
		v1.y*v2.z - v1.z*v2.y,
		v1.z*v2.x - v1.x*v2.z,
		v1.x*v2.y - v1.y*v2.x,
	}
}

func P2P(v1, v2 Vec3) Vec3 {
	return Vec3{v2.x - v1.x, v2.y - v1.y, v2.z - v1.z}
}

func SumVec3(v1, v2 Vec3) Vec3 {
	return Vec3{v1.x + v2.x, v1.y + v2.y, v1.z + v2.z}
}

type Ray struct {
	ro, rd Vec3
}

type Camera struct {
	Pos, LookAt Vec3
	v, u        Vec3
	Fov         float64
}

func NewCamera(Pos, LookAt Vec3, Fov float64) *Camera {
	ro := Pos
	n := P2P(ro, LookAt)
	d := n.len()
	n.normalize()
	cos1 := n.dot(Vec3{0, -1, 0})
	h := d / cos1
	p2 := Vec3{ro.x, ro.y - h, ro.z}
	v := P2P(p2, LookAt)
	v.normalize()
	u := Cross(n, v)
	fmt.Println(u)
	return &Camera{Pos: Pos, LookAt: LookAt, v: v, u: u, Fov: Fov}
}

func castRay(ro, rd *Vec3, resT *float64) bool {
	var dt float64 = 0.02
	var mint float64 = 1
	var maxt float64 = 50 //10
	var lh float64 = 0
	var ly float64 = 0
	for t := mint; t < maxt; t += dt {
		p := SumVec3(*ro, rd.multiply(t))
		h := f(p.x, p.z)
		if p.y < h {
			*resT = t - dt + dt*(lh-ly)/(p.y-ly-h+lh)
			return true
		}
		dt = t * 0.02
		lh = h
		ly = p.y
	}

	return false
}

func renderImage(img *image.NRGBA, c *Camera) {
	yres := float64(height)
	xres := float64(width)
	ro := c.Pos
	p := c.LookAt
	n := P2P(ro, p)
	d := n.len()
	x := 2 * d * math.Tan((c.Fov*(math.Pi)/180)/2)
	y := x * (yres / xres)
	dx := x / xres
	dy := y / yres
	mx := xres / 2
	my := yres / 2

	for j := 0; j < height; j++ {
		if j%10 == 0 {
			fmt.Println(float64(j)/float64(height)*100, "%")
		}
		for i := 0; i < width; i++ {
			pixelcenter := SumVec3(p, SumVec3(c.u.multiply((float64(i)+0.5-mx)*dx), c.v.multiply((float64(j)+0.5-my)*dy)))
			rd := P2P(ro, pixelcenter)
			rd.normalize()
			var t float64
			if castRay(&ro, &rd, &t) {
				l := Vec3{-4, 4, 1}
				l.normalize()
				q := SumVec3(ro, rd.multiply(t)) //Pos

				k := 0.0000001
				fy := f(q.x, q.z)
				df1 := fy - f(q.x+k, q.z)
				df2 := fy - f(q.x, q.z+k)
				s := Vec3{-df1 / k, 1, -df2 / k}
				//	s = Vec3{math.Cos(q.x) * math.Cos(q.z), 1, math.Sin(q.x) * math.Sin(q.z)}

				s.normalize()
				dot := s.dot(l)
				fo := math.Exp(-0.02 * t) //fo := math.Exp(-0.08 * t)
				dot = scale(dot, -1, 1, 0, 1) * fo
				n := (1 - fo) * 80

				img.Set(i, height-j, color.NRGBA{
					R: uint8(150*dot + n),
					G: uint8(100*dot + n),
					B: uint8(100*dot + n),
					A: 255,
				})

			} else {
				/*
					img.Set(i, height-j, color.NRGBA{
						R: 70,
						G: 70,
						B: 100,
						A: 255,
					})
				*/
			}
		}
	}
}

var p *perlin.Perlin

func init() {
	p = perlin.NewPerlin(2, 2, 3, 111)
}

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	o := Vec3{0, 4, 0}  //050
	at := Vec3{6, 2, 6} //323
	d := P2P(o, at)
	c := NewCamera(SumVec3(o, d.multiply(-0.6)), at, 90) //-0.6
	renderImage(img, c)
	/*
		for j := 0; j < height; j++ {
			for i := 0; i < width; i++ {
				var total float64
				t := float64(1)
				var th float64
				x := float64(i) / float64(width)
				y := float64(j) / float64(height)
				for i := 0; i < 5; i++ {
					c := p.Noise2D((x*math.Cos(th)-y*math.Sin(th))*t, (x*math.Sin(th)+y*math.Cos(th))*t) / t
					total += c
					t *= 2
					th += 0.3
				}
				total = scale(total, -1, 1, 0, 1)
				n := uint8(total * 255)
				img.Set(i, height-j, color.NRGBA{
					R: n,
					G: n,
					B: n,
					A: 255,
				})
			}
		}*/
	f, err := os.Create("image.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func f(x, z float64) float64 {

	x /= 10
	z /= 10
	var total float64
	t := float64(1)
	var th float64
	for i := 0; i < 10; i++ {
		c := p.Noise2D((x*math.Cos(th)-z*math.Sin(th))*t, (x*math.Sin(th)+z*math.Cos(th))*t) / t
		total += c
		t *= 1.99
		th += 0.3
	}
	total = scale(total, -1, 1, 0, 1)
	return total * 5.8

}

func scale(number, inMin, inMax, outMin, outMax float64) float64 {
	return (number-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
