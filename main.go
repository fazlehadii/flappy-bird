package main

import (
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	bird       Bird
	pipes      []Pipe

	started bool
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

type Bird struct {
	imgOptions    ebiten.DrawImageOptions
	frameIdx      int
	rotation      float64
	gravity       float64
	velocityY     float64
	rotationSpeed float64
	x             float64
	y             float64
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

	birdDownImage, _, birdDownImageError := ebitenutil.NewImageFromFile("assets/images/bird_down.png")
	if birdDownImageError != nil {
		log.Fatal(birdDownImageError)
	}
	birdMidImage, _, birdMidImageError := ebitenutil.NewImageFromFile("assets/images/bird_mid.png")
	if birdMidImageError != nil {
		log.Fatal(birdMidImageError)
	}
	birdUpImage, _, birdUpImageError := ebitenutil.NewImageFromFile("assets/images/bird_up.png")
	if birdUpImageError != nil {
		log.Fatal(birdUpImageError)
	}

	startImage, _, startImageError := ebitenutil.NewImageFromFile("assets/images/start.png")
	if startImageError != nil {
		log.Fatal(startImageError)
	}

	g.assets = Assets{
		background: backroundImage,
		ground:     groundImage,
		start:      startImage,
		pipeTop:    pipeTopImage,
		pipeBottom: pipeBottomImage,
		birdDown:   birdDownImage,
		birdMid:    birdMidImage,
		birdUp:     birdUpImage,
	}

}

func NewBird() Bird {
	b := Bird{
		imgOptions:    ebiten.DrawImageOptions{},
		frameIdx:      0,
		x:             (float64(screenWidth) / 2) - 50,
		y:             250,
		gravity:       0.5,
		velocityY:     -7.5,
		rotationSpeed: 0,
	}

	b.imgOptions.GeoM.SetElement(0, 2, b.x)
	b.imgOptions.GeoM.SetElement(1, 2, b.y)

	return b
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

func (b *Bird) jump() {
	b.velocityY = -7.5
	b.rotationSpeed = 0
}

func (b *Bird) update(g *Game) {

	if b.velocityY < 0 {
		b.y += b.velocityY
		b.velocityY += b.gravity

		if b.rotation > -35 {
			b.rotation -= 5

		}

	} else if b.velocityY >= 0 {

		if b.rotation < 45 {
			b.rotation += b.rotationSpeed

			if b.rotationSpeed <= 4.7 {
				b.rotationSpeed += 0.4

			}
		}

		if b.velocityY < 10 {
			b.velocityY += b.gravity
		}

		b.y += b.velocityY
	}

	w, h := g.assets.birdMid.Bounds().Dx(), g.assets.birdMid.Bounds().Dy()

	b.imgOptions.GeoM.Reset()
	b.imgOptions.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	b.imgOptions.GeoM.Rotate(b.rotation * math.Pi / 180)
	b.imgOptions.GeoM.Translate(b.x, b.y)
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
	p.x -= 2.5

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
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {

		if !g.started {
			g.started = true
			g.pipes = append(g.pipes, NewPipe(g))
		}

		g.bird.jump()

	}

	g.background.update()
	g.ground.update()

	if ebiten.Tick()%5 == 0 {
		g.bird.frameIdx = (g.bird.frameIdx + 1) % 3
	}

	if !g.started {
		return nil
	}

	g.bird.update(g)

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

	if g.started {

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
	}

	screen.DrawImage(g.assets.ground, &g.ground.img1options)
	screen.DrawImage(g.assets.ground, &g.ground.img2options)

	switch g.bird.frameIdx {
	case 0:
		screen.DrawImage(g.assets.birdDown, &g.bird.imgOptions)

	case 1:
		screen.DrawImage(g.assets.birdMid, &g.bird.imgOptions)

	case 2:
		screen.DrawImage(g.assets.birdUp, &g.bird.imgOptions)
	}

	if !g.started {
		startMsgImgOptions := ebiten.DrawImageOptions{}
		w, h := g.assets.start.Bounds().Dx(), g.assets.start.Bounds().Dy()

		startMsgImgOptions.GeoM.SetElement(0, 2, (float64(screenWidth)/2)-(float64(w)/2))
		startMsgImgOptions.GeoM.SetElement(1, 2, (float64(screenHeight)/2)-(float64(h)/2))

		screen.DrawImage(g.assets.start, &startMsgImgOptions)
	}

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
	g.bird = NewBird()

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
