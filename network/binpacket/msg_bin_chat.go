package binpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgBinChat is a binpacket for chat messages.
type MsgBinChat struct {
	Unk0 uint8
	Type uint8
	Flags uint16
	Message string
	SenderName string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgBinChat) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

// Parse parses the packet from binary
func (m *MsgBinChat) Parse(bf *byteframe.ByteFrame) error {
	m.Unk0 = bf.ReadUint8()
	m.Type = bf.ReadUint8()
	m.Flags = bf.ReadUint16()
	senderNameSize := bf.ReadUint16()
	messageSize := bf.ReadUint16()

	// TODO(Andoryuuta): Need proper shift-jis and null termination.
	m.Message = string(bf.ReadBytes(uint(messageSize))[:messageSize-1])
	m.SenderName = string(bf.ReadBytes(uint(senderNameSize))[:senderNameSize-1])

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgBinChat) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(m.Unk0)
	bf.WriteUint8(m.Type)
	bf.WriteUint16(m.Flags)
	bf.WriteUint16(uint16(len(m.SenderName)+1))
	bf.WriteUint16(uint16(len(m.Message)+1))
	bf.WriteNullTerminatedBytes([]byte(m.Message))
	bf.WriteNullTerminatedBytes([]byte(m.SenderName))

	return nil
}
