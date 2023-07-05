package mcdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"time"

	"phoenixbuilder/fastbuilder/lib/minecraft/mirror"
	"phoenixbuilder/minecraft/nbt"

	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/chunk"

	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
)

// Provider implements a world provider for the Minecraft world format, which is based on a leveldb database.
type Provider struct {
	DB  *leveldb.DB
	dir string
	D   data
}

// chunkVersion is the current version of chunks.
const chunkVersion = 27

// New creates a new provider reading and writing from/to files under the path passed. If a world is present
// at the path, New will parse its data and initialise the world with it. If the data cannot be parsed, an
// error is returned.
// A compression type may be passed which will be used for the compression of new blocks written to the database. This
// will only influence the compression. Decompression of the database will happen based on IDs found in the compressed
// blocks.
var ErrCannotOpenMCDB = errors.New("cannot open mc database")

func New(dir string, compression opt.Compression, readOnly bool, dbOptions *opt.Options) (*Provider, error) {
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	p := &Provider{dir: dir}
	initData := func() (err error) {
		// A level.dat was not currently present for the world.
		p.initDefaultLevelDat()
		return p.saveAuxInfo()
	}
	openData := func() (err error) {
		f, err := ioutil.ReadFile(filepath.Join(dir, "level.dat"))
		if err != nil {
			return fmt.Errorf("error opening level.dat file: %w", err)
		}
		// The first 8 bytes are a useless header (version and length): We don't need it.
		if len(f) < 8 {
			// The file did not have enough content, meaning it is corrupted. We return an error.
			return fmt.Errorf("level.dat exists but has no data")
		}
		if err := nbt.UnmarshalEncoding(f[8:], &p.D, nbt.LittleEndian); err != nil {
			return fmt.Errorf("error decoding level.dat NBT: %w", err)
		}
		return nil
	}

	if _, err := os.Stat(filepath.Join(dir, "level.dat")); os.IsNotExist(err) {
		err = initData()
		if err != nil {
			return nil, err
		}
	} else {
		err = openData()
		if err != nil {
			err = initData()
			if err != nil {
				return nil, err
			}
		}
	}

	if dbOptions == nil {
		dbOptions = &opt.Options{
			Compression: compression,
			BlockSize:   16 * opt.KiB,
		}
		if readOnly {
			dbOptions.ReadOnly = true
		}
	}

	if db, err := leveldb.OpenFile(filepath.Join(dir, "db"), dbOptions); err != nil {
		return nil, fmt.Errorf("error opening leveldb database: %w", err)
	} else {
		p.DB = db
	}

	return p, nil
}

// initDefaultLevelDat initialises a default level.dat file.
func (p *Provider) initDefaultLevelDat() {
	p.D.DoDayLightCycle = true
	p.D.DoWeatherCycle = true
	p.D.BaseGameVersion = GameVersion
	p.D.NetworkVersion = int32(NetworkVersion)
	p.D.LastOpenedWithVersion = minimumCompatibleClientVersion
	p.D.MinimumCompatibleClientVersion = minimumCompatibleClientVersion
	p.D.LevelName = "World"
	p.D.GameType = 1 // creative
	p.D.StorageVersion = 8
	p.D.Generator = 1
	p.D.Abilities.WalkSpeed = 0.1
	p.D.Abilities.AttackMobs = true
	p.D.Abilities.AttackPlayers = true
	p.D.Abilities.Mine = true
	p.D.Abilities.DoorsAndSwitches = true
	p.D.Abilities.FlySpeed = 0.05
	p.D.Abilities.Flying = false
	p.D.Abilities.InstantBuild = true
	p.D.Abilities.Mine = true
	p.D.Abilities.OpenContainers = true
	p.D.Abilities.Teleport = true
	p.D.Abilities.WalkSpeed = 0.1
	p.D.Abilities.OP = true
	p.D.PVP = false
	p.D.WorldStartCount = 1
	p.D.RandomTickSpeed = 1
	p.D.FallDamage = true
	p.D.FireDamage = true
	p.D.DrowningDamage = true
	p.D.CommandsEnabled = true
	p.D.MultiPlayerGame = true
	p.D.SpawnX = 0
	p.D.SpawnZ = 0
	p.D.SpawnY = 330
	p.D.ShowCoordinates = true
	p.D.Difficulty = 1 //peaceful
	p.D.DoWeatherCycle = true
	p.D.RainLevel = 1.0
	p.D.LightningLevel = 1.0
	p.D.ServerChunkTickRange = 6
	p.D.NetherScale = 8
}

// LoadChunk loads a chunk at the position passed from the leveldb database. If it doesn't exist, exists is
// false. If an error is returned, exists is always assumed to be true.
func (p *Provider) loadChunk(position define.ChunkPos) (c *chunk.Chunk, exists bool, err error) {
	data := chunk.SerialisedData{}
	key := p.index(position)

	// This key is where the version of a chunk resides. The chunk version has changed many times, without any
	// actual substantial changes, so we don't check this.
	_, err = p.DB.Get(append(key, keyVersion), nil)
	if err == leveldb.ErrNotFound {
		// The new key was not found, so we try the old key.
		if _, err = p.DB.Get(append(key, keyVersionOld), nil); err != nil {
			return nil, false, nil
		}
	} else if err != nil {
		return nil, true, fmt.Errorf("error reading version: %w", err)
	}
	data.SubChunks = make([][]byte, (define.WorldRange.Height()>>4)+1)
	for i := range data.SubChunks {
		data.SubChunks[i], err = p.DB.Get(append(key, keySubChunkData, uint8(i+(define.WorldRange[0]>>4))), nil)
		if err == leveldb.ErrNotFound {
			// No sub chunk present at this Y level. We skip this one and move to the next, which might still
			// be present.
			continue
		} else if err != nil {
			return nil, true, fmt.Errorf("error reading sub chunk data %v: %w", i, err)
		}
	}
	c, err = chunk.DiskDecode(data, define.WorldRange)
	return c, true, err
}

// LoadBlockNBT loads all block entities from the chunk position passed.
func (p *Provider) LoadBlockNBT(position define.ChunkPos) ([]map[string]any, error) {
	data, err := p.DB.Get(append(p.index(position), keyBlockEntities), nil)
	if err != leveldb.ErrNotFound && err != nil {
		return nil, err
	}
	var a []map[string]any

	buf := bytes.NewBuffer(data)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	for buf.Len() != 0 {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block NBT: %w", err)
		}
		a = append(a, m)
	}
	return a, nil
}

func (p *Provider) loadTimeStamp(position define.ChunkPos) (timeStamp int64) {
	data, err := p.DB.Get(append(p.index(position), keyTimeStamp), nil)
	if err != nil || len(data) != 0 {
		return mirror.TimeStampNotFound
	}
	return int64(binary.LittleEndian.Uint64(data))
}

func (o *Provider) GetWithNoFallBack(pos define.ChunkPos) (data *mirror.ChunkData) {
	return o.Get(pos)
}

func (p *Provider) Get(pos define.ChunkPos) (data *mirror.ChunkData) {
	cd := &mirror.ChunkData{}
	cd.ChunkPos = pos
	if c, found, err := p.loadChunk(pos); err != nil || !found {
		return nil
	} else {
		cd.Chunk = c
	}
	if nbts, err := p.LoadBlockNBT(pos); err == nil {
		cd.BlockNbts = make(map[define.CubePos]map[string]interface{})
		for _, nbt := range nbts {
			if pos, success := define.GetCubePosFromNBT(nbt); success {
				cd.BlockNbts[pos] = nbt
			}
		}
	}
	cd.SyncTime = p.loadTimeStamp(pos)
	return cd
}

func (p *Provider) GetWithDeadline(pos define.ChunkPos, deadline time.Time) (data *mirror.ChunkData) {
	return p.Get(pos)
}

// SaveChunk saves a chunk at the position passed to the leveldb database. Its version is written as the
// version in the chunkVersion constant.
func (p *Provider) SaveChunk(position define.ChunkPos, c *chunk.Chunk) error {
	data := chunk.Encode(c, chunk.DiskEncoding)

	key := p.index(position)
	_ = p.DB.Put(append(key, keyVersion), []byte{chunkVersion}, nil)
	// Write the heightmap by just writing 512 empty bytes.
	//_ = p.db.Put(append(key, key3DData), append(make([]byte, 512), data.Biomes...), nil)

	finalisation := make([]byte, 4)
	binary.LittleEndian.PutUint32(finalisation, 2)
	_ = p.DB.Put(append(key, keyFinalisation), finalisation, nil)

	for i, sub := range data.SubChunks {
		_ = p.DB.Put(append(key, keySubChunkData, byte(i+(c.Range()[0]>>4))), sub, nil)
	}
	return nil
}

// SaveBlockNBT saves all block NBT data to the chunk position passed.
func (p *Provider) SaveBlockNBT(position define.ChunkPos, data []map[string]any) error {
	if len(data) == 0 {
		return p.DB.Delete(append(p.index(position), keyBlockEntities), nil)
	}
	buf := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, d := range data {
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("error encoding block NBT: %w", err)
		}
	}
	return p.DB.Put(append(p.index(position), keyBlockEntities), buf.Bytes(), nil)
}

func (p *Provider) saveTimeStamp(position define.ChunkPos, timeStamp int64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(timeStamp))
	return p.DB.Put(append(p.index(position), keyTimeStamp), buf, nil)
}

func (p *Provider) Write(cd *mirror.ChunkData) error {
	if err := p.SaveChunk(cd.ChunkPos, cd.Chunk); err != nil {
		return err
	}
	serializedNbt := make([]map[string]interface{}, 0)
	for _, nbt := range cd.BlockNbts {
		serializedNbt = append(serializedNbt, nbt)
	}
	if err := p.SaveBlockNBT(cd.ChunkPos, serializedNbt); err != nil {
		return err
	}
	if err := p.saveTimeStamp(cd.ChunkPos, cd.SyncTime); err != nil {
		return err
	}
	return nil
}

type ChunksInfo struct {
	hasTimeStamp       bool
	hasVersion         bool
	hasNbtKey          bool
	haskeyFinalisation bool
	SubChunksCount     uint8
}

func (p *Provider) IterAll() map[define.ChunkPos]*ChunksInfo {
	iter := p.DB.NewIterator(nil, nil)
	result := make(map[define.ChunkPos]*ChunksInfo)
	var r *ChunksInfo
	var hasK bool
	for iter.Next() {
		key := iter.Key()
		if len(key) < 8 || len(key) > 10 {
			fmt.Println(string(key))
			continue
		}
		pos, rest_key := p.Position(key)
		if len(rest_key) > 0 {
			if r, hasK = result[pos]; !hasK {
				r = &ChunksInfo{}
				result[pos] = r
			}
			switch rest_key[0] {
			case keySubChunkData:
				r.SubChunksCount++
			case keyTimeStamp:
				r.hasTimeStamp = true
			case keyVersion:
				r.hasVersion = true
			case keyBlockEntities:
				r.hasNbtKey = true
			case keyFinalisation:
				r.haskeyFinalisation = true
			}
		}
	}
	return result
}

func (p *Provider) saveAuxInfo() (err error) {
	p.D.LastPlayed = time.Now().Unix()
	f, err := os.OpenFile(filepath.Join(p.dir, "level.dat"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening level.dat file: %w", err)
	}
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, int32(9))
	nbtData, err := nbt.MarshalEncoding(p.D, nbt.LittleEndian)
	if err != nil {
		return fmt.Errorf("error encoding level.dat to NBT: %w", err)
	}
	_ = binary.Write(buf, binary.LittleEndian, int32(len(nbtData)))
	_, _ = buf.Write(nbtData)

	_, _ = f.Write(buf.Bytes())

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing level.dat: %w", err)
	}
	//noinspection SpellCheckingInspection
	if err := os.WriteFile(filepath.Join(p.dir, "levelname.txt"), []byte(p.D.LevelName), 0644); err != nil {
		return fmt.Errorf("error writing levelname.txt: %w", err)
	}
	return nil
}

// Close closes the provider, saving any file that might need to be saved, such as the level.dat.
func (p *Provider) Close(readOnly bool) error {
	// p.initDefaultLevelDat()
	if !readOnly {
		p.saveAuxInfo()
	}
	return p.DB.Close()
}

// index returns a byte buffer holding the written index of the chunk position passed. If the dimension passed to New
// is not world.Overworld, the length of the index returned is 12. It is 8 otherwise.
func (p *Provider) index(position define.ChunkPos) []byte {
	x, z, dim := uint32(position[0]), uint32(position[1]), uint32(WorldDimension)
	b := make([]byte, 12)

	binary.LittleEndian.PutUint32(b, x)
	binary.LittleEndian.PutUint32(b[4:], z)
	if dim == 0 {
		return b[:8]
	}
	binary.LittleEndian.PutUint32(b[8:], dim)
	return b
}
func (p *Provider) Position(key []byte) (position define.ChunkPos, rest []byte) {
	x := binary.LittleEndian.Uint32(key[0:4])
	z := binary.LittleEndian.Uint32(key[4:8])
	if len(key) == 8 {
		return define.ChunkPos{int32(x), int32(z)}, key[8:]
	}
	if key[8] < 3 {
		return define.ChunkPos{int32(x), int32(z)}, key[12:]
	}
	return define.ChunkPos{int32(x), int32(z)}, key[8:]
}
