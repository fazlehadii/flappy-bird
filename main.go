package main

import (
	// "fmt"

	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const screenWidth int = 551
const screenHeight int = 670

type Assets struct {
	background,
	ground,

	pipeTop,
	pipeBottom,

	birdDown,
	birdMid,
	birdUp,

	gameOver,
	start *ebiten.Image
}

type Game struct {
	assets     Assets
	background Background
	ground     Ground
	pipes      []Pipe
}

type Background struct {
	img1options ebiten.DrawImageOptions
	img1cords   [2]float64

	img2options ebiten.DrawImageOptions
	img2cords   [2]float64
}

type Ground struct {
	img1options ebiten.DrawImageOptions
	img1cords   [2]float64

	img2options ebiten.DrawImageOptions
	img2cords   [2]float64
}

type Pipe struct {
	pipeType   int
	imgOptions ebiten.DrawImageOptions
	x          float64
	y          float64
}

const (
	pipeTop    int = 0
	pipeBottom int = 1
	pipeBoth   int = 2
)

func LoadAssets(g *Game) {
	backroundImage, _, backgroundImageError := ebitenutil.NewImageFromFile("assets/images/background.png")
	if backgroundImageError != nil {
		log.Fatal(backgroundImageError)
	}

	groundImage, _, groundImageError := ebitenutil.NewImageFromFile("assets/images/ground.png")
	if groundImageError != nil {
		log.Fatal(groundImageError)
	}

	pipeTopImage, _, pipeTopImageError := ebitenutil.NewImageFromFile("assets/images/pipe_top.png")
	if pipeTopImageError != nil {
		log.Fatal(pipeTopImageError)
	}
	pipeBottomImage, _, pipeBottomImageError := ebitenutil.NewImageFromFile("assets/images/pipe_bottom.png")
	if pipeBottomImageError != nil {
		log.Fatal(pipeBottomImageError)
	}

	g.assets = Assets{
		background: backroundImage,
		ground:     groundImage,
		pipeTop:    pipeTopImage,
		pipeBottom: pipeBottomImage,
	}

}

func NewPipe(g *Game) Pipe {
	pipeType := rand.IntN(100)

	switch {
	case pipeType < 70:
		pipeType = 2
	case pipeType < 85:
		pipeType = 1
	default:
		pipeType = 0
	}

	p := Pipe{
		pipeType:   pipeType,
		imgOptions: ebiten.DrawImageOptions{},
		x:          (rand.Float64()*(200-150) + 150) + float64(screenWidth),
	}

	p.imgOptions.GeoM.SetElement(0, 2, p.x)

	switch pipeType {
	case pipeTop:
		p.y = (rand.Float64()*(float64(screenHeight)-300-200) + 200) - 800

		p.imgOptions.GeoM.SetElement(1, 2, p.y)

	case pipeBottom:
		p.y = rand.Float64()*(float64(screenHeight)-200-300) + 300

		p.imgOptions.GeoM.SetElement(1, 2, p.y)

	case pipeBoth:
		p.y = (rand.Float64()*(float64(screenHeight)-300-200) + 200) - 800

		p.imgOptions.GeoM.SetElement(1, 2, p.y)

	}

	return p
}

func NewBackground() Background {
	b := Background{
		img1options: ebiten.DrawImageOptions{},
		img2options: ebiten.DrawImageOptions{},
		img1cords:   [2]float64{0, 0},
		img2cords:   [2]float64{float64(screenWidth), 0},
	}

	b.img1options.GeoM.SetElement(1, 2, b.img1cords[1])
	b.img2options.GeoM.SetElement(1, 2, b.img2cords[1])

	return b
}

func NewGround() Ground {
	g := Ground{
		img1options: ebiten.DrawImageOptions{},
		img2options: ebiten.DrawImageOptions{},
		img1cords:   [2]float64{0, float64(screenHeight) - 160},
		img2cords:   [2]float64{float64(screenWidth), float64(screenHeight) - 160},
	}

	g.img1options.GeoM.SetElement(1, 2, g.img1cords[1])
	g.img2options.GeoM.SetElement(1, 2, g.img2cords[1])

	return g
}

func (b *Background) update() {
	b.img1cords[0] -= 1
	b.img2cords[0] -= 1

	b.img1options.GeoM.SetElement(0, 2, b.img1cords[0])
	b.img2options.GeoM.SetElement(0, 2, b.img2cords[0])

	if b.img2cords[0] <= 0 {
		b.img1cords[0] = 0
		b.img2cords[0] = float64(screenWidth)
	}
}

func (g *Ground) update() {
	g.img1cords[0] -= 2.5
	g.img2cords[0] -= 2.5

	g.img1options.GeoM.SetElement(0, 2, g.img1cords[0])
	g.img2options.GeoM.SetElement(0, 2, g.img2cords[0])

	if g.img2cords[0] <= 0 {
		g.img1cords[0] = 0
		g.img2cords[0] = float64(screenWidth)
	}
}

func (p *Pipe) update() {
	p.x -= 5.5

	switch p.pipeType {
	case pipeTop:
		p.imgOptions.GeoM.SetElement(0, 2, p.x)

	case pipeBottom:
		p.imgOptions.GeoM.SetElement(0, 2, p.x)

	case pipeBoth:
		p.imgOptions.GeoM.SetElement(0, 2, p.x)
	}

}

func (g *Game) Update() error {
	g.background.update()
	g.ground.update()

	for i := range g.pipes {
		g.pipes[i].update()
	}

	if g.pipes[len(g.pipes)-1].x <= (float64(screenWidth) - 55) {
		g.pipes = append(g.pipes, NewPipe(g))
	}

	if g.pipes[0].x <= -60 {
		g.pipes = g.pipes[1:]
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.DrawImage(g.assets.background, &g.background.img1options)
	screen.DrawImage(g.assets.background, &g.background.img2options)

	for _, pipe := range g.pipes {
		switch pipe.pipeType {
		case pipeTop:
			screen.DrawImage(g.assets.pipeTop, &pipe.imgOptions)
		case pipeBottom:
			screen.DrawImage(g.assets.pipeBottom, &pipe.imgOptions)
		case pipeBoth:
			screen.DrawImage(g.assets.pipeTop, &pipe.imgOptions)
			pipe.imgOptions.GeoM.SetElement(1, 2, (pipe.y+800)+120)
			screen.DrawImage(g.assets.pipeBottom, &pipe.imgOptions)
		}

	}

	screen.DrawImage(g.assets.ground, &g.ground.img1options)
	screen.DrawImage(g.assets.ground, &g.ground.img2options)

}

func (g *Game) Layout(_ int, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := Game{}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Flappy Bird by Maliks")

	LoadAssets(&g)

	g.background = NewBackground()
	g.ground = NewGround()
	g.pipes = append(g.pipes, NewPipe(&g))

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
