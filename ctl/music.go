package ctl

import (
	"github.com/bolg-developers/MikanMusic-API/model"
	"github.com/bolg-developers/MikanMusic-API/svc"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
)

func CreateMusic(c *gin.Context) {
	var mr model.MusicAndGenre
	if err := c.BindJSON(&mr); err != nil {
		c.Status(400)
		log.Printf("BadRequest: %+v", errors.WithStack(err))
		return
	}

	if err := svc.ValidateGenres(mr.Genres); err != nil {
		if err == svc.ErrUnknownGenre {
			c.Status(400)
			log.Printf("BadRequest: %+v", err)
			return
		}
		c.Status(500)
		log.Printf("InternalServerError: %+v", err)
		return
	}

	if mr.ArtworkURL == "" {
		mr.ArtworkURL = model.IconDefaultURL
	}

	if err := svc.CreateMusic(&mr.Music); err != nil {
		c.Status(500)
		log.Printf("InternalServerError: %+v", err)
		return
	}

	for _, g := range mr.Genres {
		if err := svc.CreateMusicGenreRelation(&model.MusicGenreRelation{
			MusicID: mr.ID,
			GenreID: g.ID,
		}); err != nil {
			log.Printf("InternalServerError: %+v", err)
			if err := svc.DeleteMusic(mr.ID); err != nil {
				log.Printf("InternalServerError: %+v", err)
			}
			if err := svc.DeleteMusicGenreRelationByMusicID(mr.ID); err != nil {
				log.Printf("InternalServerError: %+v", err)
			}
			c.Status(500)
			return
		}
	}

	c.Status(201)
}

// MusicGenreの一覧を返す
func GetAllMusics(c *gin.Context) {
	musics, err := svc.GetAllMusics()
	if err != nil {
		c.Status(500)
		log.Printf("InternalServerError: %+v", err)
		return
	}
	mgList := make(model.MusicAndGenreList, 0)

	for _, music := range musics {
		genres, err := svc.GetGenresByMusicID(music.ID)
		if err != nil {
			c.Status(500)
			log.Printf("InternalServerError: %+v", err)
			return
		}
		mgList = append(mgList, &model.MusicAndGenre{
			Music:  *music,
			Genres: genres,
		})
	}
	c.JSON(200, gin.H{"musics": mgList})
}

func IncrementMusicCntListen(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Status(404)
		return
	}

	m, err := svc.GetMusicByID(id)
	if err != nil {
		c.Status(400)
		log.Printf("BadRequest: %+v", err)
		return
	}

	m.CountListen++

	if err := svc.UpdateMusic(m); err != nil {
		c.Status(500)
		log.Printf("InternalServerError: %+v", err)
		return
	}

	c.Status(200)
}
