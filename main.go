package main

import (
	"image"
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

	birdMasks  [3]Mask
	pipeMasks  [2]Mask
	groundMask Mask
}

type Mask struct {
	width  int
	height int
	bits   []byte
}

type GameState int

const (
	StateReady GameState = iota
	StatePlaying
	StateGameOver
)

type Game struct {
	assets     Assets
	background Background
	ground     Ground
	bird       Bird
	pipes      []Pipe

	started bool
	state   GameState
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

func MakeMask(img image.Image) Mask {
	var bits []byte
	var currentByte byte
	var bitCount uint

	addPixel := func(solid bool) {
		if solid {
			currentByte |= 1 << bitCount
		}

		bitCount++

		if bitCount == 8 {
			bits = append(bits, currentByte)
			currentByte = 0
			bitCount = 0
		}

	}

	m := Mask{
		width:  img.Bounds().Dx(),
		height: img.Bounds().Dy(),
		bits:   []byte{},
	}

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			_, _, _, a := img.At(x, y).RGBA()

			addPixel(a >= 32768)
		}
	}

	if bitCount > 0 {
		bits = append(bits, currentByte)
	}

	m.bits = bits

	return m
}

func LoadAssets(g *Game) {
	backroundImage, _, backgroundImageError := ebitenutil.NewImageFromFile("assets/images/background.png")
	if backgroundImageError != nil {
		log.Fatal(backgroundImageError)
	}

	groundImage, orgGroundImage, groundImageError := ebitenutil.NewImageFromFile("assets/images/ground.png")
	if groundImageError != nil {
		log.Fatal(groundImageError)
	}

	pipeTopImage, orgPipeTopImage, pipeTopImageError := ebitenutil.NewImageFromFile("assets/images/pipe_top.png")
	if pipeTopImageError != nil {
		log.Fatal(pipeTopImageError)
	}
	pipeBottomImage, orgPipeBottomImage, pipeBottomImageError := ebitenutil.NewImageFromFile("assets/images/pipe_bottom.png")
	if pipeBottomImageError != nil {
		log.Fatal(pipeBottomImageError)
	}

	birdDownImage, orgBirdDownImage, birdDownImageError := ebitenutil.NewImageFromFile("assets/images/bird_down.png")
	if birdDownImageError != nil {
		log.Fatal(birdDownImageError)
	}
	birdMidImage, orgBirdMidImage, birdMidImageError := ebitenutil.NewImageFromFile("assets/images/bird_mid.png")
	if birdMidImageError != nil {
		log.Fatal(birdMidImageError)
	}
	birdUpImage, orgBirdUpImage, birdUpImageError := ebitenutil.NewImageFromFile("assets/images/bird_up.png")
	if birdUpImageError != nil {
		log.Fatal(birdUpImageError)
	}

	startImage, _, startImageError := ebitenutil.NewImageFromFile("assets/images/start.png")
	if startImageError != nil {
		log.Fatal(startImageError)
	}
	gameOverImage, _, gameOverImageError := ebitenutil.NewImageFromFile("assets/images/game_over.png")
	if gameOverImageError != nil {
		log.Fatal(gameOverImageError)
	}

	g.assets = Assets{
		background: backroundImage,
		ground:     groundImage,
		start:      startImage,
		gameOver:   gameOverImage,
		pipeTop:    pipeTopImage,
		pipeBottom: pipeBottomImage,
		birdDown:   birdDownImage,
		birdMid:    birdMidImage,
		birdUp:     birdUpImage,
	}

	g.assets.groundMask = MakeMask(orgGroundImage)
	g.assets.birdMasks[0] = MakeMask(orgBirdDownImage)
	g.assets.birdMasks[1] = MakeMask(orgBirdMidImage)
	g.assets.birdMasks[2] = MakeMask(orgBirdUpImage)
	g.assets.pipeMasks[0] = MakeMask(orgPipeTopImage)
	g.assets.pipeMasks[1] = MakeMask(orgPipeBottomImage)

}

func CheckCollisions(g *Game) bool {
	birdMask := g.assets.birdMasks[g.bird.frameIdx]

	halfW := float64(birdMask.width) / 2
	halfH := float64(birdMask.height) / 2

	angle := g.bird.rotation * math.Pi / 180

	cos := math.Abs(math.Cos(angle))
	sin := math.Abs(math.Sin(angle))

	rotatedHalfW := halfW*cos + halfH*sin
	rotatedHalfH := halfW*sin + halfH*cos

	birdRect := image.Rect(
		int(math.Floor(g.bird.x-rotatedHalfW)),
		int(math.Floor(g.bird.y-rotatedHalfH)),
		int(math.Ceil(g.bird.x+rotatedHalfW)),
		int(math.Ceil(g.bird.y+rotatedHalfH)),
	)
	var pipeRect image.Rectangle

	for _, pipe := range g.pipes {
		if pipe.x <= float64(birdRect.Max.X) &&
			pipe.x+52 >= float64(birdRect.Min.X) {

			switch pipe.pipeType {
			case pipeBoth:
				pipeRect = image.Rect(
					int(pipe.x), int(pipe.y),
					int(pipe.x+52), int(pipe.y+800),
				)

				if birdRect.Overlaps(pipeRect) {
					if MaskOverlap(
						birdMask, g.assets.pipeMasks[0],
						g.bird.x, g.bird.y,
						pipe.x, pipe.y,
						g.bird.rotation,
					) {
						return true
					}
				}

				bottomY := pipe.y + 800 + 120

				pipeRect = image.Rect(
					int(pipe.x), int(bottomY),
					int(pipe.x+52), int(bottomY+800),
				)

				if birdRect.Overlaps(pipeRect) {
					if MaskOverlap(
						birdMask, g.assets.pipeMasks[1],
						g.bird.x, g.bird.y,
						pipe.x, bottomY,
						g.bird.rotation,
					) {
						return true
					}
				}

			default:
				pipeRect = image.Rect(
					int(pipe.x), int(pipe.y),
					int(pipe.x+52), int(pipe.y+800),
				)

				if birdRect.Overlaps(pipeRect) {
					if MaskOverlap(
						birdMask, g.assets.pipeMasks[pipe.pipeType],
						g.bird.x, g.bird.y,
						pipe.x, pipe.y,
						g.bird.rotation,
					) {
						return true
					}
				}
			}
		}
	}

	groundY := g.ground.img1cords[1]

	groundRect := image.Rect(
		0, int(groundY),
		screenWidth, int(groundY+219),
	)

	if birdRect.Overlaps(groundRect) {
		if MaskOverlap(
			birdMask, g.assets.groundMask,
			g.bird.x, g.bird.y,
			0, groundY,
			g.bird.rotation,
		) {
			return true
		}
	}

	return false
}

func MaskOverlap(
	birdMask, obstacleMask Mask,
	birdCenterX, birdCenterY float64,
	obstacleX, obstacleY float64,
	rotation float64,
) bool {

	halfW := float64(birdMask.width) / 2
	halfH := float64(birdMask.height) / 2

	angle := rotation * math.Pi / 180

	cos := math.Cos(angle)
	sin := math.Sin(angle)

	// Rotated bird's AABB
	rotatedHalfW := halfW*math.Abs(cos) + halfH*math.Abs(sin)
	rotatedHalfH := halfW*math.Abs(sin) + halfH*math.Abs(cos)

	// Intersection between rotated bird AABB and obstacle
	x0 := math.Max(birdCenterX-rotatedHalfW, obstacleX)
	y0 := math.Max(birdCenterY-rotatedHalfH, obstacleY)

	x1 := math.Min(birdCenterX+rotatedHalfW, obstacleX+float64(obstacleMask.width))
	y1 := math.Min(birdCenterY+rotatedHalfH, obstacleY+float64(obstacleMask.height))

	if x0 >= x1 || y0 >= y1 {
		return false
	}

	// Iterate over screen pixels in the intersection
	for worldY := int(math.Floor(y0)); worldY < int(math.Ceil(y1)); worldY++ {
		for worldX := int(math.Floor(x0)); worldX < int(math.Ceil(x1)); worldX++ {

			// Centre of the current screen pixel
			screenX := float64(worldX) + 0.5
			screenY := float64(worldY) + 0.5

			if screenX < x0 || screenX >= x1 ||
				screenY < y0 || screenY >= y1 {
				continue
			}

			// Position relative to bird's centre
			dx := screenX - birdCenterX
			dy := screenY - birdCenterY

			// Inverse rotation
			birdLocalX := cos*dx + sin*dy + halfW
			birdLocalY := -sin*dx + cos*dy + halfH

			birdPixelX := int(math.Floor(birdLocalX))
			birdPixelY := int(math.Floor(birdLocalY))

			// Obstacle coordinates (no rotation)
			obstaclePixelX := int(math.Floor(screenX - obstacleX))
			obstaclePixelY := int(math.Floor(screenY - obstacleY))

			// Make sure both coordinates are inside their masks
			if birdPixelX < 0 || birdPixelX >= birdMask.width ||
				birdPixelY < 0 || birdPixelY >= birdMask.height {
				continue
			}

			if obstaclePixelX < 0 || obstaclePixelX >= obstacleMask.width ||
				obstaclePixelY < 0 || obstaclePixelY >= obstacleMask.height {
				continue
			}

			// Flatten coordinates
			birdPixelIdx := birdPixelY*birdMask.width + birdPixelX
			obstaclePixelIdx := obstaclePixelY*obstacleMask.width + obstaclePixelX

			// Locate bits
			birdByteIdx := birdPixelIdx / 8
			birdBitIdx := birdPixelIdx % 8

			obstacleByteIdx := obstaclePixelIdx / 8
			obstacleBitIdx := obstaclePixelIdx % 8

			// Retrieve pixels
			birdSolid := birdMask.bits[birdByteIdx]&(1<<birdBitIdx) != 0
			obstacleSolid := obstacleMask.bits[obstacleByteIdx]&(1<<obstacleBitIdx) != 0

			if birdSolid && obstacleSolid {
				return true
			}
		}
	}

	return false
}

func NewBird() Bird {
	b := Bird{
		imgOptions: ebiten.DrawImageOptions{},
		frameIdx:   0,
		x:          (float64(screenWidth) / 2) - 50,
		y:          250,
		gravity:    0.4,
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
	case pipeBoth:
		p.y = (rand.Float64()*(float64(screenHeight)-300-200) + 200) - 800

		p.imgOptions.GeoM.SetElement(1, 2, p.y)

	case pipeTop:
		p.y = (rand.Float64()*(float64(screenHeight)-300-200) + 200) - 800

		p.imgOptions.GeoM.SetElement(1, 2, p.y)

	case pipeBottom:
		p.y = rand.Float64()*(float64(screenHeight)-200-300) + 300

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
	b.velocityY = -6
	b.rotationSpeed = 0
}

func (b *Bird) update(g *Game) {

	if b.velocityY < 0 {
		b.y += b.velocityY
		if b.y-12 <= 0 {
			b.y = 12
			b.velocityY = 0
		}

		b.velocityY += b.gravity

		if b.rotation > -30 {
			b.rotation -= 10

		}

	} else if b.velocityY >= 0 {

		if b.rotation < 35 {
			b.rotation += b.rotationSpeed

			if b.rotationSpeed <= 4.7 {
				b.rotationSpeed += 0.4

			}
		}

		if b.velocityY < 10 {
			b.velocityY += b.gravity
		}

		b.y += b.velocityY
		if b.y-12 <= 0 {
			b.y = 12
			b.velocityY = 0
		}

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

func (g *Game) reset() {
	g.background = NewBackground()
	g.ground = NewGround()
	g.bird = NewBird()

	g.pipes = nil
	g.state = StateReady
}

func (g *Game) Update() error {
	if g.state == StateGameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.reset()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {

		if g.state == StateReady {
			g.state = StatePlaying
			g.pipes = append(g.pipes, NewPipe(g))
		}

		g.bird.jump()
	}

	g.background.update()
	g.ground.update()

	if ebiten.Tick()%5 == 0 {
		g.bird.frameIdx = (g.bird.frameIdx + 1) % 3
	}

	if g.state != StatePlaying {
		return nil
	}

	g.bird.update(g)

	for i := range g.pipes {
		g.pipes[i].update()
	}

	if CheckCollisions(g) {
		g.state = StateGameOver
		return nil
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

	if g.state == StatePlaying || g.state == StateGameOver {

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

	if g.state == StateReady {
		startMsgImgOptions := ebiten.DrawImageOptions{}
		w, h := g.assets.start.Bounds().Dx(), g.assets.start.Bounds().Dy()

		startMsgImgOptions.GeoM.SetElement(0, 2, (float64(screenWidth)/2)-(float64(w)/2))
		startMsgImgOptions.GeoM.SetElement(1, 2, (float64(screenHeight)/2)-(float64(h)/2))

		screen.DrawImage(g.assets.start, &startMsgImgOptions)
	}

	if g.state == StateGameOver {
		w, h := g.assets.gameOver.Bounds().Dx(), g.assets.gameOver.Bounds().Dy()

		var options ebiten.DrawImageOptions
		options.GeoM.Translate(
			float64(screenWidth-w)/2,
			float64(screenHeight-h)/2,
		)

		screen.DrawImage(g.assets.gameOver, &options)
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
