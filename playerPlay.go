package main

import "time"
import "github.com/Tnze/go-mc/net/packet"
import "fmt"
import "bytes"
import "github.com/Tnze/go-mc/nbt"

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

func (player *Player) handlePlaying() {
	player.sendJoinGame()
}

func (player *Player) sendJoinGame() {
	var dimCodec dimensionCodec
	var buf bytes.Buffer
	
	err := nbt.Marshal(&buf, dimCodec);
	if err != nil {
		fmt.Println("Could Not Marshal NBT Data! Error: ", err)
		return
	}
	
	err = player.connection.WritePacket(packet.Marshal(0x24,
		packet.Int(0),                        // entity id
		packet.Boolean(false),                // is hardcore
		packet.UnsignedByte(3),               // gamemode
		packet.UnsignedByte(3),               // "previous gamemode"
		packet.VarInt(0),                     // world count
		packet.String("minecraft:coolworld"), // world names(array)
		packet.NBT(buf),                         // dimension codec
//		packet.NBT(),                         // dimention
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
