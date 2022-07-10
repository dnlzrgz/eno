package chip8

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// 	 +---------------+= 0xFFF (4095) End Chip-8 RAM
// 	 |               |
// 	 |               |
// 	 |               |
// 	 |               |
// 	 |               |
// 	 | 0x200 to 0xFFF|
// 	 |     Chip-8    |
// 	 | Program / Data|
// 	 |     Space     |
// 	 |               |
// 	 |               |
// 	 |               |
// 	 +- - - - - - - -+= 0x600 (1536) Start ETI 660 Chip-8 programs
// 	 |               |
// 	 |               |
// 	 |               |
// 	 +---------------+= 0x200 (512) Start of most Chip-8 programs
// 	 | 0x000 to 0x1FF|
// 	 | Reserved for  |
// 	 |  interpreter  |
// 	 +---------------+= 0x000 (0) Begin Chip-8 RAM. We store font data here instead of storing the interpreter because we don't have that restriction.
//

const maxRomSize = 0xFFF - 0x200

// represents the chip-8 virtual machine.
type vm struct {

	// the chip-8 memory had 4kb, which corresponds to 4096 bytes.
	memory [4096]byte

	// registers exist as a kind of short-term memory and the chip-8 had 16 8-bit
	// registers. They're referred to as V0 through VF.
	registers [16]byte

	// chip-8 had the ability to go into subroutines, and a stack for keeping track
	// of where to return to. The stack is 16 16-bit values, meaning the program can
	// go into 16 nested subroutines.
	stack [16]uint16

	// is used to store return addresses when calling procedures
	sp uint16

	// Timers, both 8-bit registers for deciding when to beep.
	st byte
	dt byte

	// It's a special 16-bit register that exists mostly for reading and writing to
	// memory in general.
	i uint16

	// The program counter (pc) stores the address of the current instruction as an
	// 16-bit integer. Every single instruction in chip-8 will update the program
	// counter when it's done to go to the next instruction.
	// In the chip-8 memory layout, 0x000 to 0x1FF memory is reserved. That's why
	// it starts at 0x200
	pc uint16

	// Opcode under examination
	opcode uint16

	// CPU clock
	clock *time.Ticker

	*video

	keypad [16]uint8

	// draw flag signals when to update the screen
	drawFlag bool

	// channel for sending or receiving audio signals
	audioCh chan struct{}

	// channel for sending or receiving shutdown signals
	ShutdownCh chan struct{}
}

// Loads into memory the content of the rom specified by the path. If the rom is
// too large or there is a problem while reading it returns an error.
func (v *vm) loadROM(path string) error {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(rom) > maxRomSize {
		return fmt.Errorf("the room size is to large")
	}

	for i := 0; i < len(rom); i++ {
		v.memory[0x200+i] = rom[i] // write to memory with pc offset
	}

	return nil
}

// Loads into memory the font set into the first 80 bytes of memory.
func (v *vm) loadFontSet() {
	for i := 0; i < len(fontSet); i++ {
		v.memory[i] = fontSet[i]
	}
}

// NewVm initializes and returns a new chip-8 instance. It returns an error if
// there is any problem while reading the rom or loading it into memory or if the creation
//of the video output fails.
func NewVm(path string, clockSpeed int) (*vm, error) {
	video, err := newVideo()
	if err != nil {
		return nil, err
	}

	vm := &vm{
		memory:     [4096]byte{},
		registers:  [16]byte{},
		stack:      [16]uint16{},
		pc:         0x200,
		clock:      time.NewTicker(time.Second / time.Duration(clockSpeed)),
		video:      video,
		keypad:     [16]byte{},
		audioCh:    make(chan struct{}),
		ShutdownCh: make(chan struct{}),
	}

	vm.loadFontSet()
	if err := vm.loadROM(path); err != nil {
		return nil, err
	}

	return vm, nil
}

// Run starts the emulator and with a clock that runs by default at 60MHz
func (v *vm) Run() {
	for {
		select {
		case <-v.clock.C:
			if !v.window.Closed() {
				v.cycle()
				v.update()
				v.handleKeyInput()
				v.delayTimerTick()
				v.soundTimerTick()
				continue
			}
			break
		case <-v.ShutdownCh:
			break
		}
		break
	}

	v.shutdown("shutting down...")
}

func (v *vm) cycle() {
	v.opcode = uint16(v.memory[v.pc])<<8 | uint16(v.memory[v.pc+1])
	v.drawFlag = false

	if err := v.parseOpcode(); err != nil {
		fmt.Printf("error parsing opcode: %v", err)
	}
}

func (v *vm) update() {
	if v.drawFlag {
		v.video.draw()
		return
	}

	v.video.window.UpdateInput()
}

func (v *vm) handleKeyInput() {
	for i, key := range v.video.keyMap {
		if v.video.window.JustReleased(key) && v.video.keysDown[i] != nil {
			v.video.keysDown[i].Stop()
			v.video.keysDown[i] = nil
		} else if v.window.JustPressed(key) {
			if v.video.keysDown[i] == nil {
				v.video.keysDown[i] = time.NewTicker(time.Second / 5)
			}

			v.keypad[i] = byte(1)
		}

		if v.video.keysDown[i] == nil {
			continue
		}

		select {
		case <-v.video.keysDown[i].C:
			v.keypad[i] = byte(1)
		default:
		}
	}
}

func (v *vm) delayTimerTick() {
	if v.dt > 0 {
		v.dt--
	}
}

func (v *vm) soundTimerTick() {
	if v.st > 0 {
		if v.st == 1 {
			v.audioCh <- struct{}{}
		}

		v.st--
	}
}

func (v *vm) shutdown(msg string) {
	log.Println(msg)
	close(v.audioCh)
	v.ShutdownCh <- struct{}{}
}

func (v *vm) drawSprite(x, y uint16) {
	height := v.opcode & 0x000F
	v.memory[0xF] = 0

	for i := uint16(0); i < height; i++ {
		p := uint16(v.memory[v.i+i])

		for j := uint16(0); j < 8; j++ {
			index := x + j + ((y + i) * 64)
			if index >= uint16(len(v.video.gfx)) {
				continue
			}

			if (p & (0x80 >> j)) != 0 {
				if v.video.gfx[index] == 1 {
					v.registers[0xF] = 1
				}

				v.video.gfx[index] ^= 1
			}
		}
	}

	v.drawFlag = true
}
