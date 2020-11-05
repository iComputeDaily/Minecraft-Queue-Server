package main

import "time"
import "github.com/Tnze/go-mc/net/packet"
import "fmt"
import "bytes"
import "github.com/Tnze/go-mc/nbt"
// import "errors"

type nbtWrap struct {
	Value interface{} // Holds struct of nbt data
	buf *bytes.Buffer // Holds encoded nbt
	enc *nbt.Encoder  // An encoder object
}

type dimensionCodec struct {
	DimensionType struct {
		Type string       `nbt:"type"`
		Value []dimension `nbt:"value"`
	} `nbt:"minecraft:dimension_type"`
	
	WorldgenBiome struct {
		Type string       `nbt:"type"`
		Value []biome     `nbt:"value"`
	} `nbt:"minecraft:worldgen/biome"`
}

type dimension struct {
	Name string `nbt:"name"`
	Id int      `nbt:"id"` // BUG(iComputeDaily): might be another number type not shure
	Element struct {
		PiglinSafe byte         `nbt:"piglin_safe"`
		Natural byte            `nbt:"natural"`
		AmbientLight float32    `nbt:"ambient_light"`
		FixedTime int64         `nbt:"fixed_time"`
		Infiniburn string       `nbt:"infiniburn"`
		RespawnAnchorWorks byte `nbt:"respawn_anchor_works"`
 		HasSkylight byte        `nbt:"has_skylight"`
 		BedWorks byte           `nbt:"bed_works"`
 		Effects string          `nbt:"effects"`
 		HasRaids byte           `nbt:"has_raids"`
 		LogicalHeight int       `nbt:"logical_height"`
 		CoordinateScale float32 `nbt:"coordinate_scale"`
 		Ultrawarm byte          `nbt:"ultrawarm"`
 		HasCeiling byte         `nbt:"has_ceiling"`
	} `nbt:"element"`
}

type biome struct {
	
}

func (nbtData nbtWrap) Encode() []byte {
	if nbtData.buf == nil {
		// make an empty buffer
		nbtData.buf = &bytes.Buffer{}
	}
	
	if nbtData.enc == nil {
		// make an encoder on the buffer
		nbtData.enc = nbt.NewEncoder(nbtData.buf)
	}
	
	// clear the buffer - allows memory to be reused
	nbtData.buf.Reset()
	
	// encode to the buffer
	err := nbtData.enc.Encode(nbtData.Value)
	
	if err != nil {
		panic(fmt.Sprintln("Failed To Encode! Error: ", err))
	}
	
	// get bytes from the buffer
	return nbtData.buf.Bytes()
}

func (player *Player) handlePlaying() {
	player.sendJoinGame()
}

func (player *Player) sendJoinGame() {
	var nbtDimCodec nbtWrap
	var dimCodec dimensionCodec
	nbtDimCodec.Value = dimCodec
	
	fmt.Printf("\n%+v\n\n", nbtDimCodec)
	
	err := player.connection.WritePacket(packet.Marshal(0x24,
		packet.Int(0),                        // entity id
		packet.Boolean(false),                // is hardcore
		packet.UnsignedByte(3),               // gamemode
		packet.UnsignedByte(3),               // "previous gamemode"
		packet.VarInt(0),                     // world count
		packet.String("minecraft:coolworld"), // world names(array)
		nbtDimCodec,                          // dimension codec
//		nbtDim,                               // dimention
		packet.String("minecraft:coolworld"), // world name
		packet.Long(0),                       // hashed seed
		packet.VarInt(1),                     // max players
		packet.VarInt(2),                     // veiw distance
		packet.Boolean(false),                // reduced debug info
		packet.Boolean(false),                // enable respawn screen
		packet.Boolean(false),                // is debug
		packet.Boolean(true),                 // is falt
		))
	if err != nil {
		fmt.Println("Could Not Send Join Game Packet! Error: ", err)
		return
	}
	
	time.Sleep(3600 * time.Second)
}
