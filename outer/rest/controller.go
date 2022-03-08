package rest

import (
	"bytes"
	"sort"

	"github.com/cockroachdb/pebble"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/qyqx233/go-tunel/outer/pub"
	"github.com/qyqx233/gtool/lib/convert"
	"github.com/rs/zerolog/log"
)

type Bytes []byte

// func (b []byte) Compare(o Bytes) int {
// 	return bytes.Compare(b, o)
// }

func (b Bytes) Compare(o Bytes) int {
	return bytes.Compare(b, o)
}

type Comparabler[T any] interface {
	Compare(o T) int
}

func SearchSliceFunc(n int, fx func(int) int) int {
	l, h := 0, n-1
	for l <= h {
		m := int(uint(l+h) >> 1) // avoid overflow when computing h
		r := fx(m)
		if r > 0 {
			h = m - 1 // preserves f(i-1) == false
		} else if r < 0 {
			l = m + 1 // preserves f(j) == true
		} else {
			return m
		}
	}
	return -1
}

func SearchComparableSlice[T Comparabler[T]](sl []T, t T) int {
	l, h := 0, len(sl)-1
	for l <= h {
		m := int(uint(l+h) >> 1) // avoid overflow when computing h
		r := sl[m].Compare(t)
		if r > 0 {
			h = m - 1 // preserves f(i-1) == false
		} else if r < 0 {
			l = m + 1 // preserves f(j) == true
		} else {
			return m
		}
	}
	return -1
}

func SearchSlice[T string | ~int](sl []T, t T) int {
	l, h := 0, len(sl)-1
	for l <= h {
		m := int(uint(l+h) >> 1) // avoid overflow when computing h
		// println(l, h, m)
		// i ≤ h < j
		if sl[m] > t {
			h = m - 1 // preserves f(i-1) == false
		} else if sl[m] < t {
			l = m + 1 // preserves f(j) == true
		} else {
			return m
		}
	}
	return -1
}

func StartRest(addr string) {
	log.Debug().Str("addr", addr).Msg("监听地址")
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))
	app.Get("/api/memStor", func(c *fiber.Ctx) error {
		return c.JSON(&pub.MemStor)
	})
	app.Get("/api/table/:kind", func(c *fiber.Ctx) error {
		type Msg struct {
			Hosts []string `json:"hosts"`
		}
		var msg = new(Msg)
		var key = []byte("hosts:")
		oldValue, _ := PebbleGetBytes(key)
		var hosts [][]byte
		if len(oldValue) > 0 {
			hosts = bytes.Split(oldValue, []byte(":"))
		}
		strs := make([]string, 0, 10)
		for _, host := range hosts {
			strs = append(strs, convert.Bytes2String(host))
		}
		msg.Hosts = strs
		return c.JSON(msg)
	})
	app.Delete("/api/table/:kind", func(c *fiber.Ctx) error {
		var key []byte
		var kind = c.Params("kind")
		if kind == "hosts" {
			key = []byte("hosts:")
		}
		pdb.Delete(key, pebble.Sync)
		return c.JSON(&Response{Code: 0})
	})
	app.Post("/api/table/:kind", func(c *fiber.Ctx) error {
		type HostSt struct {
			Hosts []string `json:"hosts"`
		}
		var msg = new(HostSt)
		var kind = c.Params("kind")
		if kind == "hosts" {
			c.BodyParser(msg)
			var key = []byte("hosts:")
			var hosts = [][]byte{}
			oldValue, _ := PebbleGetBytes(key)
			if len(oldValue) > 0 {
				hosts = bytes.Split(oldValue, []byte(":"))
			}
			for _, host := range msg.Hosts {
				var sb = convert.String2Bytes(host)
				index := SearchSliceFunc(len(hosts), func(i int) int {
					return bytes.Compare(hosts[i], sb)
				})
				if index >= 0 {
					continue
				}
				hosts = append(hosts, sb)
			}
			sort.Slice(hosts, func(i, j int) bool {
				return bytes.Compare(hosts[i], hosts[j]) < 0
			})
			var value = bytes.Join(hosts, []byte(":"))
			pdb.Set(key, value, pebble.Sync)
		}
		return nil
	})
	app.Post("/api/transport", func(c *fiber.Ctx) error {
		type Msg struct {
			Port   int  `json:"port"`
			Enable bool `json:"enable"`
		}
		var msg = new(Msg)
		c.BodyParser(msg)
		if msg.Port == 0 {
			return c.JSON(&Response{500, "port empty"})
		}
		var tsp = TransportPdb{
			Enable: msg.Enable,
		}
		// ts := pub.GetConfig().GetTs(msg.Port)
		var key []byte
		key = append(key, "port:"...)
		key = append(key, convert.Uint642Bytes(uint64(msg.Port))[0])
		ts := pub.GetConfig().GetTs(msg.Port)
		if t, ok := pub.MemStor.Transports[msg.Port]; ok {
			t.Enable = msg.Enable
			log.Debug().Interface("MemStor.Transports", t).Interface("xx", pub.MemStor.Transports).Msg("print")
		}
		ts.Enable = msg.Enable
		pdb.Set(key, tsp.encode(), pebble.Sync)
		return c.JSON(&Response{200, "success"})
	})
	app.Get("/api/transport", func(c *fiber.Ctx) error {
		var response = new(ListTransport)
		var config = pub.GetConfig()
		var ts TransportPdb
		for _, t := range config.Transport {
			if !t.Export {
				continue
			}
			var key = make([]byte, 0, 32)
			var enable = true
			key = append(key, "port:"...)
			key = append(key, convert.Uint642Bytes(uint64(t.LocalPort))[0])
			err := PebbleGet(key, &ts)
			if err == nil {
				enable = ts.Enable
			}
			response.Data = append(response.Data, Transport{
				TargetHost: t.TargetHost,
				TargetPort: t.TargetPort,
				Port:       t.LocalPort,
				Enable:     enable,
			})
		}
		return c.JSON(response)
		// return nil
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello")
	})

	err := app.Listen(addr)
	if err != nil {
		log.Error().Err(err).Msg("fiber listen error")
	}
}
