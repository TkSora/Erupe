package channelserver

import (
	"fmt"
	"strings"

	"github.com/Andoryuuta/Erupe/network/binpacket"
	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
)

// MSG_SYS_CAST[ED]_BINARY types enum
const (
	BinaryMessageTypeState = 0
	BinaryMessageTypeChat  = 1
	BinaryMessageTypeEmote = 6
)

// MSG_SYS_CAST[ED]_BINARY broadcast types enum
const (
	BroadcastTypeTargeted = 0x01
	BroadcastTypeStage    = 0x03
	BroadcastTypeWorld    = 0x0a
)

func sendServerChatMessage(s *Session, message string) {
	// Make the inside of the casted binary
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0: 0,
		Type: 5,
		Flags: 0x80,
		Message: message,
		SenderName: "Erupe",
	}
	msgBinChat.Build(bf)

	castedBin := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.charID,
		Type0:          0,
		Type1:          BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	s.QueueSendMHF(castedBin)
}

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)

	// Parse out the real casted binary payload
	var realPayload []byte
	var msgBinTargeted *binpacket.MsgBinTargeted
	if pkt.Type0 == BroadcastTypeTargeted {
		bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
		msgBinTargeted = &binpacket.MsgBinTargeted{}
		err := msgBinTargeted.Parse(bf)

		if err != nil {
			s.logger.Warn("Failed to parse targeted cast binary")
			return
		}

		realPayload = msgBinTargeted.RawDataPayload
	} else {
		realPayload = pkt.RawDataPayload
	}

	// Make the response to forward to the other client(s).
	resp := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.charID,
		Type0:          pkt.Type0, // (The client never uses Type0 upon receiving)
		Type1:          pkt.Type1,
		RawDataPayload: realPayload,
	}

	// Send to the proper recipients.
	switch pkt.Type0 {
	case BroadcastTypeWorld:
		s.server.BroadcastMHF(resp, s)
	case BroadcastTypeStage:
		s.stage.BroadcastMHF(resp, s)
	case BroadcastTypeTargeted:
		for _, targetID := range (*msgBinTargeted).TargetCharIDs {
			char := s.server.FindSessionByCharID(targetID)

			if char != nil {
				char.QueueSendMHF(resp)
			}
		}
	default:
		s.stage.BroadcastMHF(resp, s)
	}

	// Handle chat commands
	if pkt.Type1 == BinaryMessageTypeChat {
		bf := byteframe.NewByteFrameFromBytes(realPayload)

		// IMPORTANT! Casted binary objects are sent _as they are in memory_,
		// this means little endian for LE CPUs, might be different for PS3/PS4/PSP/XBOX.
		bf.SetLE()

		chatMessage := &binpacket.MsgBinChat{}
		chatMessage.Parse(bf)

		fmt.Printf("Got chat message: %+v\n", chatMessage)


		if strings.HasPrefix(chatMessage.Message, "!tele") {
			var x, y int16
			n, err := fmt.Sscanf(chatMessage.Message, "!tele %d %d", &x, &y)
			if err != nil || n != 2 {
				sendServerChatMessage(s, "Invalid command. Usage:\"!tele 500 500\"")
			} else {
				sendServerChatMessage(s, fmt.Sprintf("Teleporting to %f %f", x, y))

				// Make the inside of the casted binary
				payload := byteframe.NewByteFrame()
				payload.WriteUint8(2) // SetState type(position == 2)
				payload.WriteInt16(x) // X
				payload.WriteInt16(y) // Y
				payloadBytes := payload.Data()

				posUpdate := &mhfpacket.MsgSysCastedBinary{
					CharID:         s.charID,
					Type0:          1,
					Type1:          0,
					RawDataPayload: payloadBytes,
				}

				s.QueueSendMHF(posUpdate)
			}
		}

		


		



		/*
			// Made the inside of the casted binary
			payload := byteframe.NewByteFrame()
			payload.WriteUint16(uint16(i)) // Chat type

			//Chat type 0 = World
			//Chat type 1 = Local
			//Chat type 2 = Guild
			//Chat type 3 = Alliance
			//Chat type 4 = Party
			//Chat type 5 = Whisper
			//Thanks to @Alice on discord for identifying these.

			payload.WriteUint8(0) // Unknown
			msg := fmt.Sprintf("Chat type %d", i)
			playername := fmt.Sprintf("Ando")
			payload.WriteUint16(uint16(len(playername) + 1))
			payload.WriteUint16(uint16(len(msg) + 1))
			payload.WriteUint8(0) // Is this correct, or do I have the endianess of the prev 2 fields wrong?
			payload.WriteNullTerminatedBytes([]byte(msg))
			payload.WriteNullTerminatedBytes([]byte(playername))
			payloadBytes := payload.Data()

			//Wrap it in a CASTED_BINARY packet to broadcast
			bfw := byteframe.NewByteFrame()
			bfw.WriteUint16(uint16(network.MSG_SYS_CASTED_BINARY))
			bfw.WriteUint32(0x23325A29) // Character ID
			bfw.WriteUint8(1)           // type
			bfw.WriteUint8(1)           // type2
			bfw.WriteUint16(uint16(len(payloadBytes)))
			bfw.WriteBytes(payloadBytes)
		*/
	}
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}
