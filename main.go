package main

import (
	// "fmt"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const screenWidth int = 551
const screenHeight int = 670

type Game struct {
	background Background
	ground     Ground
}

type Background struct {
	backgroundImage *ebiten.Image

	backgroundImage1Options ebiten.DrawImageOptions
	backgroundImage1Cords   [2]float64

	backgroundImage2Options ebiten.DrawImageOptions
	backgroundImage2Cords   [2]float64
}

type Ground struct {
	groundImage *ebiten.Image

	groundImage1Options ebiten.DrawImageOptions
	groundImage1Cords   [2]float64

	groundImage2Options ebiten.DrawImageOptions
	groundImage2Cords   [2]float64
}

func NewBackground(img *ebiten.Image) Background {
	return Background{
		backgroundImage: img,

		backgroundImage1Options: ebiten.DrawImageOptions{},
		backgroundImage2Options: ebiten.DrawImageOptions{},

		backgroundImage1Cords: [2]float64{0, 0},
		backgroundImage2Cords: [2]float64{551, 0},
	}
}

func NewGround(img *ebiten.Image) Ground {
	return Ground{
		groundImage: img,

		groundImage1Options: ebiten.DrawImageOptions{},
		groundImage2Options: ebiten.DrawImageOptions{},

		groundImage1Cords: [2]float64{0, float64(screenHeight) - 160},
		groundImage2Cords: [2]float64{551, float64(screenHeight) - 160},
	}
}

func (g *Game) Update() error {
	g.background.backgroundImage1Cords[0] -= 1
	g.background.backgroundImage2Cords[0] -= 1

	g.ground.groundImage1Cords[0] -= 2.5
	g.ground.groundImage2Cords[0] -= 2.5

	g.background.backgroundImage1Options.GeoM.SetElement(0, 2, g.background.backgroundImage1Cords[0])
	g.background.backgroundImage1Options.GeoM.SetElement(1, 2, g.background.backgroundImage1Cords[1])

	g.background.backgroundImage2Options.GeoM.SetElement(0, 2, g.background.backgroundImage2Cords[0])
	g.background.backgroundImage2Options.GeoM.SetElement(1, 2, g.background.backgroundImage2Cords[1])

	g.ground.groundImage1Options.GeoM.SetElement(0, 2, g.ground.groundImage1Cords[0])
	g.ground.groundImage1Options.GeoM.SetElement(1, 2, g.ground.groundImage1Cords[1])

	g.ground.groundImage2Options.GeoM.SetElement(0, 2, g.ground.groundImage2Cords[0])
	g.ground.groundImage2Options.GeoM.SetElement(1, 2, g.ground.groundImage2Cords[1])

	if g.ground.groundImage2Cords[0] <= 0 {
		g.ground.groundImage1Cords[0] = 0
		g.ground.groundImage2Cords[0] = 551
	}

	if g.background.backgroundImage2Cords[0] <= 0 {
		g.background.backgroundImage1Cords[0] = 0
		g.background.backgroundImage2Cords[0] = 551
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.DrawImage(g.background.backgroundImage, &g.background.backgroundImage1Options)
	screen.DrawImage(g.background.backgroundImage, &g.background.backgroundImage2Options)

	screen.DrawImage(g.ground.groundImage, &g.ground.groundImage1Options)
	screen.DrawImage(g.ground.groundImage, &g.ground.groundImage2Options)
}

func (g *Game) Layout(_ int, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := Game{}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Flappy Bird by Maliks")

	backroundImage, _, backgroundImageError := ebitenutil.NewImageFromFile("assets/images/background.png")
	if backgroundImageError != nil {
		log.Fatal(backgroundImageError)
	}

	groundImage, _, groundImageError := ebitenutil.NewImageFromFile("assets/images/ground.png")
	if groundImageError != nil {
		log.Fatal(groundImageError)
	}

	g.background = NewBackground(backroundImage)
	g.ground = NewGround(groundImage)

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
