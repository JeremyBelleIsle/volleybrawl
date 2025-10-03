package main

import (
	"bytes"
	_ "embed"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/go-mp3"
)

//go:embed pixel-drift.mp3
var musique []byte

const (
	screenWidth  = 900
	screenHeight = 600
	playerW      = 40
	playerH      = 60
	borderW      = 24
	midLineW     = 12
	playerSpeed  = 4
	sampleRate   = 44100
)

var audioContext *audio.Context
var player *audio.Player

// Player représente un joueur simple
type Player struct {
	X, Y                                       float64
	W, H                                       float64
	Color                                      color.RGBA
	LeftKey, RightKey, UpKey, DownKey, DashKey ebiten.Key
	MinX, MaxX                                 float64
	cooldown                                   float64
	playerSpeed                                int
	posXD                                      int
	posYD                                      int
	Dead                                       bool
	deadCooldown                               float32
}

// Game contient l'état
type Game struct {
	leftBorderColor  color.RGBA
	rightBorderColor color.RGBA
	midLineColor     color.RGBA
	p1a, p1b         *Player
	p2a, p2b         *Player
	balleX           float32
	balleY           float32
	targetX          float32
	targetY          float32
	Next_targetX     float32
	Next_targetY     float32
	SpeedBalle       float32
	Win              int
}

var (
	mplusFaceSource *text.GoTextFaceSource
)

func init() {
	rand.Seed(time.Now().UnixNano())
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}
func (p *Player) Update() {
	if !p.Dead {
		if p.cooldown > 0 {
			p.cooldown -= 1.0 / 60.0
			if p.cooldown < 0 {
				p.cooldown = 0
			}
		}
		if p.cooldown > 0 {
			p.playerSpeed = 4
		}
		if ebiten.IsKeyPressed(p.LeftKey) {
			p.X -= float64(p.playerSpeed)
			if p.playerSpeed == 16 && math.Abs(float64(p.posXD)-p.X) >= 100 || math.Abs(float64(p.posYD)-p.Y) >= 100 && p.playerSpeed == 16 {
				p.cooldown = 3.0
				p.posXD = 0
				p.posYD = 0
			}
		}
		if ebiten.IsKeyPressed(p.RightKey) {
			p.X += float64(p.playerSpeed)
			if p.playerSpeed == 16 && math.Abs(float64(p.posXD)-p.X) >= 100 || math.Abs(float64(p.posYD)-p.Y) >= 100 && p.playerSpeed == 16 {
				p.cooldown = 3.0
				p.posXD = 0
				p.posYD = 0
			}
		}
		if ebiten.IsKeyPressed(p.UpKey) {
			p.Y -= float64(p.playerSpeed)
			if p.playerSpeed == 16 && math.Abs(float64(p.posXD)-p.X) >= 100 || math.Abs(float64(p.posYD)-p.Y) >= 100 && p.playerSpeed == 16 {
				p.cooldown = 3.0
				p.posXD = 0
				p.posYD = 0
			}
		}
		if ebiten.IsKeyPressed(p.DownKey) {
			p.Y += float64(p.playerSpeed)
			if p.playerSpeed == 16 && math.Abs(float64(p.posXD)-p.X) >= 100 || math.Abs(float64(p.posYD)-p.Y) >= 100 && p.playerSpeed == 16 {
				p.cooldown = 3.0
				p.posXD = 0
				p.posYD = 0
			}
		}
		if ebiten.IsKeyPressed(p.DashKey) {
			if p.cooldown <= 0 && p.playerSpeed == 4 {
				p.playerSpeed = 16
				p.cooldown = 0
				p.posXD = int(p.X)
				p.posYD = int(p.Y)
			}
		}
		// Limites horizontales
		if p.X < p.MinX {
			p.X = p.MinX
		}
		if p.X+40 > screenWidth {
			p.X = screenWidth - 40
		}
		// Limites verticales
		if p.Y < 0 {
			p.Y = 0
		}
		if p.Y > float64(screenHeight)-p.H {
			p.Y = float64(screenHeight) - p.H
		}
	} else {
		p.deadCooldown -= 1.0 / 60.0
		if p.deadCooldown <= 0 {
			p.Dead = false
		}
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	if !p.Dead {
		vector.DrawFilledRect(screen, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), p.Color, true)
	}
}

func NewGame() *Game {
	g := &Game{}
	g.balleX = 451
	g.balleY = 310
	g.targetX = 100
	g.targetY = 300
	g.Win = 0
	g.SpeedBalle = 3
	g.leftBorderColor = color.RGBA{R: 70, G: 130, B: 180, A: 255}  // steelblue
	g.rightBorderColor = color.RGBA{R: 180, G: 70, B: 130, A: 255} // rose
	g.midLineColor = color.RGBA{R: 30, G: 30, B: 30, A: 255}

	// positions initiales
	p1a := &Player{
		X:           float64(100),
		Y:           float64(screenHeight - 100),
		W:           playerW,
		H:           playerH,
		Color:       color.RGBA{R: 255, G: 200, B: 0, A: 255},
		LeftKey:     ebiten.KeyA,
		RightKey:    ebiten.KeyD,
		UpKey:       ebiten.KeyW,
		DownKey:     ebiten.KeyS,
		DashKey:     ebiten.KeyE,
		MinX:        borderW,
		MaxX:        screenWidth/2 - midLineW/2,
		playerSpeed: playerSpeed,
	}
	p1b := &Player{
		X:           float64(100),
		Y:           float64(screenHeight - 200),
		W:           playerW,
		H:           playerH,
		Color:       color.RGBA{R: 255, G: 220, B: 80, A: 255},
		LeftKey:     ebiten.KeyF,
		RightKey:    ebiten.KeyH,
		UpKey:       ebiten.KeyT,
		DownKey:     ebiten.KeyG,
		DashKey:     ebiten.KeyY,
		MinX:        borderW,
		MaxX:        screenWidth/2 - midLineW/2,
		playerSpeed: playerSpeed,
	}
	p2a := &Player{
		X:           float64(screenWidth - 140),
		Y:           float64(screenHeight - 100),
		W:           playerW,
		H:           playerH,
		Color:       color.RGBA{R: 0, G: 200, B: 255, A: 255},
		LeftKey:     ebiten.KeyArrowLeft,
		RightKey:    ebiten.KeyArrowRight,
		UpKey:       ebiten.KeyArrowUp,
		DownKey:     ebiten.KeyArrowDown,
		DashKey:     ebiten.KeyShiftRight,
		MinX:        screenWidth/2 + midLineW/2,
		MaxX:        screenWidth - borderW,
		playerSpeed: playerSpeed,
	}
	p2b := &Player{
		X:           float64(screenWidth - 140),
		Y:           float64(screenHeight - 200),
		W:           playerW,
		H:           playerH,
		Color:       color.RGBA{R: 80, G: 220, B: 255, A: 255},
		LeftKey:     ebiten.KeyJ,
		RightKey:    ebiten.KeyL,
		UpKey:       ebiten.KeyI,
		DownKey:     ebiten.KeyK,
		DashKey:     ebiten.KeyO,
		MinX:        screenWidth/2 + midLineW/2,
		MaxX:        screenWidth - borderW,
		playerSpeed: playerSpeed,
	}
	g.p1a = p1a
	g.p1b = p1b
	g.p2a = p2a
	g.p2b = p2b
	return g
}
func CircleRectCollide(cx, cy, r, rx, ry, rw, rh float64) bool {
	closestX := math.Max(rx, math.Min(cx, rx+rw))
	closestY := math.Max(ry, math.Min(cy, ry+rh))
	dx := cx - closestX
	dy := cy - closestY
	return (dx*dx + dy*dy) <= (r * r)
}
func RectRectCollide(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
func (g *Game) Update() error {
	if g.Win == 0 {
		g.SpeedBalle += 0.0032
		g.p1a.Update()
		g.p1b.Update()
		g.p2a.Update()
		g.p2b.Update()
		if RectRectCollide(g.p1a.X, g.p1a.Y, playerW, playerH, g.p2a.X, g.p2a.Y, playerW, playerH) {
			if g.p1a.playerSpeed == 16 && g.p2a.cooldown <= 0 {
				g.p2a.Dead = true
				g.p2a.deadCooldown = 5.0
			}
			if g.p2a.playerSpeed == 16 && g.p1a.cooldown <= 0 {
				g.p1a.Dead = true
				g.p1a.deadCooldown = 5.0
			}
		}
		if RectRectCollide(g.p1a.X, g.p1a.Y, playerW, playerH, g.p2b.X, g.p2b.Y, playerW, playerH) {
			if g.p1a.playerSpeed == 16 && g.p2b.cooldown <= 0 {
				g.p2b.Dead = true
				g.p2b.deadCooldown = 5.0
			}
			if g.p2b.playerSpeed == 16 && g.p1a.cooldown <= 0 {
				g.p1a.Dead = true
				g.p1a.deadCooldown = 5.0
			}
		}
		if RectRectCollide(g.p1b.X, g.p1b.Y, playerW, playerH, g.p2a.X, g.p2a.Y, playerW, playerH) {
			if g.p1b.playerSpeed == 16 && g.p2a.cooldown <= 0 {
				g.p2a.Dead = true
				g.p2a.deadCooldown = 5.0
			}
			if g.p2a.playerSpeed == 16 && g.p1b.cooldown <= 0 {
				g.p1b.Dead = true
				g.p1b.deadCooldown = 5.0
			}
		}
		if RectRectCollide(g.p1b.X, g.p1b.Y, playerW, playerH, g.p2b.X, g.p2b.Y, playerW, playerH) {
			if g.p1b.playerSpeed == 16 && g.p2b.cooldown <= 0 {
				g.p2b.Dead = true
				g.p2b.deadCooldown = 5.0
			}
			if g.p2b.playerSpeed == 16 && g.p1b.cooldown <= 0 {
				g.p1b.Dead = true
				g.p1b.deadCooldown = 5.0
			}
		}
		// vecteur direction
		dx := g.targetX - g.balleX
		dy := g.targetY - g.balleY
		dist := float32(math.Hypot(float64(dx), float64(dy)))

		if dist > 30 {
			// normalisation + déplacement
			g.balleX += (dx / dist) * g.SpeedBalle
			g.balleY += (dy / dist) * g.SpeedBalle
		} else {
			if math.Abs(float64(g.targetX-g.Next_targetX)) > 500 {
				g.targetX = g.Next_targetX
				g.targetY = g.Next_targetY
			} else if CircleRectCollide(float64(g.targetX), float64(g.targetY), 30, g.p1a.X, g.p1a.Y, playerW, playerH) || CircleRectCollide(float64(g.targetX), float64(g.targetY), 30, g.p1b.X, g.p1b.Y, playerW, playerH) || CircleRectCollide(float64(g.targetX), float64(g.targetY), 30, g.p2a.X, g.p2a.Y, playerW, playerH) || CircleRectCollide(float64(g.targetX), float64(g.targetY), 30, g.p2b.X, g.p2b.Y, playerW, playerH) {
				for math.Abs(float64(g.targetX-g.Next_targetX)) < 500 {
					g.Next_targetX = float32(rand.Intn(screenWidth-200) + 100)
					g.Next_targetY = float32(rand.Intn(screenHeight-200) + 100)
					if g.Next_targetX > 300 || g.Next_targetX < 600 {
						for g.Next_targetX > 300 && g.Next_targetX < 600 {
							g.Next_targetX = float32(rand.Intn(screenWidth-200) + 100)
							g.Next_targetY = float32(rand.Intn(screenHeight-200) + 100)
						}
					}
				}
			} else {
				if g.balleX < 450 {
					g.Win = 2
				}
				if g.balleX > 450 {
					g.Win = 1
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.Win == 0 {
		// fond uni
		screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

		// bords de terrain (gauche et droite)
		vector.DrawFilledRect(screen, 0, 0, borderW, float32(screenHeight), g.leftBorderColor, true)
		vector.DrawFilledRect(screen, float32(screenWidth-borderW), 0, borderW, float32(screenHeight), g.rightBorderColor, true)

		// grosse ligne au milieu
		vector.DrawFilledRect(screen, float32(screenWidth/2-midLineW/2), 0, midLineW, float32(screenHeight), g.midLineColor, true)

		// dessiner les joueurs
		g.p1a.Draw(screen)
		g.p1b.Draw(screen)
		g.p2a.Draw(screen)
		g.p2b.Draw(screen)

		//dessiner la balle
		centerX := float32(screenWidth) / 2
		centerY := float32(screenHeight) / 2
		distToCenter := float32(math.Hypot(float64(g.balleX-centerX), float64(g.balleY-centerY)))
		maxDist := float32(math.Hypot(float64(centerX), float64(centerY)))
		minRadius := float32(35)
		maxRadius := float32(77)
		balleRadius := minRadius + (maxRadius-minRadius)*(1.0-distToCenter/maxDist)
		vector.DrawFilledCircle(screen, g.balleX, g.balleY, balleRadius, color.White, true)
		//prochain point
		vector.DrawFilledCircle(screen, g.targetX, g.targetY, 30, color.RGBA{0, 255, 0, 255}, true)
	} else {
		if g.Win == 1 {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(900/4), float64(600/2))
			op.ColorScale.ScaleWithColor(color.RGBA{222, 49, 99, 0})
			text.Draw(screen, "Team 1 WIN", &text.GoTextFace{
				Source: mplusFaceSource,
				Size:   53,
			}, op)
		}
		if g.Win == 2 {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(900/4), float64(600/2))
			op.ColorScale.ScaleWithColor(color.RGBA{222, 49, 99, 0})
			text.Draw(screen, "Team 2 WIN", &text.GoTextFace{
				Source: mplusFaceSource,
				Size:   53,
			}, op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	audioContext = audio.NewContext(sampleRate)
	decoded, err := mp3.NewDecoder(bytes.NewReader(musique))
	if err != nil {
		log.Fatal(err)
	}

	// On met la musique en boucle infinie
	player, err = audio.NewPlayer(audioContext, decoded)
	player.Rewind()
	player.Play()
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Volleybrawl")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// joueur ne peuvent pas sortirent du terrain - done
// smach
// vitesse progressive - done
// dash si touche personne personne: éliminé pendant 5 secondes dash cooldown:3
