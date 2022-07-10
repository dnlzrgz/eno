package chip8

import (
	"fmt"
	"math/rand"
)

func (v *vm) op0x00E0() {
	v.video.gfx = [64 * 32]byte{}
	v.pc += 2
}

func (v *vm) op0x00EE() {
	v.pc = v.stack[v.sp] + 2
	v.sp--
}

func (v *vm) op0x1000(nnn uint16) {
	v.pc = nnn
}

func (v *vm) op0x2000(nnn uint16) {
	v.sp++
	v.stack[v.sp] = v.pc
	v.pc = nnn
}

func (v *vm) op0x3000(x uint16, nn byte) {
	if v.registers[x] == nn {
		v.pc += 4
	} else {
		v.pc += 2
	}
}

func (v *vm) op0x4000(x uint16, nn byte) {
	if v.registers[x] != nn {
		v.pc += 4
	} else {
		v.pc += 2
	}
}

func (v *vm) op0x5000(x, y uint16) {
	if v.registers[x] == v.registers[y] {
		v.pc += 4
	} else {
		v.pc += 2
	}
}

func (v *vm) op0x6000(x uint16, nn byte) {
	v.registers[x] = nn
	v.pc += 2
}

func (v *vm) op0x7000(x uint16, nn byte) {
	v.registers[x] += nn
	v.pc += 2
}

func (v *vm) op0x0000(x, y uint16) {
	v.registers[x] = v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0001(x, y uint16) {
	v.registers[x] |= v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0002(x, y uint16) {
	v.registers[x] &= v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0003(x, y uint16) {
	v.registers[x] ^= v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0004(x, y uint16) {
	if v.registers[y] > (0xFF - v.registers[x]) {
		v.registers[0xF] = 1
	} else {
		v.registers[0xF] = 0
	}

	v.registers[x] += v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0005(x, y uint16) {
	if v.registers[y] > v.registers[x] {
		v.registers[0xF] = 0
	} else {
		v.registers[0xF] = 1
	}

	v.registers[x] -= v.registers[y]
	v.pc += 2
}

func (v *vm) op0x0006(x, y uint16) {
	v.registers[x] = v.registers[y] >> 1
	v.registers[0xF] = v.registers[y] & 0x01
	v.pc += 2
}

func (v *vm) op0x0007_1(x, y uint16) {
	if v.registers[x] > v.registers[y] {
		v.registers[0xF] = 0
	} else {
		v.registers[0xF] = 1
	}

	v.registers[x] = v.registers[y] - v.registers[x]
	v.pc += 2
}

func (v *vm) op0x000E(x, y uint16) {
	v.registers[x] = v.registers[y] << 1
	v.registers[0xF] = v.registers[y] & 0x80
	v.pc += 2
}

func (v *vm) op0x9000(x, y uint16) {
	if v.registers[x] != v.registers[y] {
		v.pc += 4
	} else {
		v.pc += 2
	}
}

func (v *vm) op0xA000(nnn uint16) {
	v.i = nnn
	v.pc += 2
}

func (v *vm) op0xB000(nnn uint16) {
	v.pc = nnn + uint16(v.registers[0])
	v.pc += 2
}

func (v *vm) op0xC000(x uint16, nn byte) {
	v.registers[x] = byte(rand.Float32()*255) & nn
	v.pc += 2
}

func (v *vm) op0xD000(x, y uint16) {
	x = uint16(v.registers[x])
	y = uint16(v.registers[y])
	v.drawSprite(x, y)
	v.pc += 2
}

func (v *vm) op0x009E(x uint16) {
	if v.keypad[v.registers[x]] == 1 {
		v.pc += 4
		v.keypad[v.registers[x]] = 0
	} else {
		v.pc += 2
	}
}

func (v *vm) op0x00A1(x uint16) {
	if v.keypad[v.registers[x]] == 0 {
		v.pc += 4
	} else {
		v.keypad[v.registers[x]] = 0
		v.pc += 2
	}
}

func (v *vm) op0x0007_2(x uint16) {
	v.registers[x] = v.dt
	v.pc += 2
}

func (v *vm) op0x000A(x uint16) {
	for i, k := range v.keypad {
		if k != 0 {
			v.registers[x] = byte(i)
			v.pc += 2
			break
		}
	}

	v.keypad[v.registers[x]] = 0
}

func (v *vm) op0x0015(x uint16) {
	v.dt = v.registers[x]
	v.pc += 2
}

func (v *vm) op0x0018(x uint16) {
	v.st = v.registers[x]
	v.pc += 2
}

func (v *vm) op0x001E(x uint16) {
	v.i += uint16(v.registers[x])
	v.pc += 2
}

func (v *vm) op0x0029(x uint16) {
	v.i = uint16(v.registers[x]) * 5
	v.pc += 2
}

func (v *vm) op0x0033(x uint16) {
	v.memory[v.i] = v.registers[x] / 100
	v.memory[v.i+1] = (v.registers[x] / 10) % 10
	v.memory[v.i+2] = (v.registers[x] % 100) % 10
	v.pc += 2
}

func (v *vm) op0x0065(x uint16) {
	for i := uint16(0); i <= x; i++ {
		v.registers[i] = v.memory[v.i+i]
	}

	v.pc += 2
}

func (v *vm) op0x0055(x uint16) {
	for i := uint16(0); i <= x; i++ {
		v.memory[v.i+i] = v.registers[i]
	}

	v.pc += 2
}

func (v *vm) unknownOp(opcode uint16) error {
	return fmt.Errorf("received an unknown opcode: %x\n", opcode)
}

func (v *vm) parseOpcode() error {
	x := (v.opcode & 0x0F00) >> 8
	y := (v.opcode & 0x00F0) >> 4
	nn := byte(v.opcode & 0x00FF)
	nnn := v.opcode & 0x0FFF

	switch v.opcode & 0xF000 {
	case 0x0000:
		switch v.opcode & 0x00FF {
		case 0x00E0:
			v.op0x00E0()
		case 0x00EE:
			v.op0x00EE()
		default:
			return v.unknownOp(v.opcode & 0x00FF)
		}
	case 0x1000:
		v.op0x1000(nnn)
	case 0x2000:
		v.op0x2000(nnn)
	case 0x3000:
		v.op0x3000(x, nn)
	case 0x4000:
		v.op0x4000(x, nn)
	case 0x5000:
		v.op0x5000(x, y)
	case 0x6000:
		v.op0x6000(x, nn)
	case 0x7000:
		v.op0x7000(x, nn)
	case 0x8000:
		switch v.opcode & 0x000F {
		case 0x0000:
			v.op0x0000(x, y)
		case 0x0001:
			v.op0x0001(x, y)
		case 0x0002:
			v.op0x0002(x, y)
		case 0x0003:
			v.op0x0003(x, y)
		case 0x0004:
			v.op0x0004(x, y)
		case 0x0005:
			v.op0x0005(x, y)
		case 0x0006:
			v.op0x0006(x, y)
		case 0x0007:
			v.op0x0007_1(x, y)
		case 0x000E:
			v.op0x000E(x, y)
		default:
			return v.unknownOp(v.opcode & 0x000F)
		}
	case 0x9000:
		v.op0x9000(x, y)
	case 0xA000:
		v.op0xA000(nnn)
	case 0xB000:
		v.op0xB000(nnn)
	case 0xC000:
		v.op0xC000(x, nn)
	case 0xD000:
		v.op0xD000(x, y)
	case 0xE000:
		switch v.opcode & 0x00FF {
		case 0x009E:
			v.op0x009E(x)
		case 0x00A1:
			v.op0x00A1(x)
		default:
			return v.unknownOp(v.opcode & 0x00FF)
		}
	case 0xF000:
		switch v.opcode & 0x00FF {
		case 0x0007:
			v.op0x0007_2(x)
		case 0x000A:
			v.op0x000A(x)
		case 0x0015:
			v.op0x0015(x)
		case 0x0018:
			v.op0x0018(x)
		case 0x001E:
			v.op0x001E(x)
		case 0x0029:
			v.op0x0029(x)
		case 0x0033:
			v.op0x0033(x)
		case 0x0055:
			v.op0x0055(x)
		case 0x0065:
			v.op0x0065(x)
		default:
			return v.unknownOp(v.opcode & 0x00FF)
		}
	default:
		return v.unknownOp(v.opcode & 0x00FF)
	}

	return nil
}
