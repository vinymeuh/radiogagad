// Copyright 2019 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package mpd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (s CurrentSong) String() string {
	return fmt.Sprintf("file: %s\nId: %d\nName: %s\nPos: %d\nTime: %d\nTitle: %s\n",
		s.File, s.ID, s.Name, s.Pos, s.Time, s.Title)
}

var currentSongTestData = CurrentSong{
	File:  "http://novazz.ice.infomaniak.ch/novazz-128.mp3",
	ID:    1,
	Name:  "novazz",
	Pos:   0,
	Time:  0,
	Title: "",
}

func (s Status) String() string {
	return fmt.Sprintf("audio: %s\nbitrate: %d\nduration: %d\nelapsed: %f\nerror: %s\nplaylist: %d\nstate: %s\nsong: %d\nsongid: %d\n",
		s.Audio, s.Bitrate, s.Duration, s.Elapsed, s.Error, s.Playlist, s.State, s.Song, s.Songid)
}

var statusTestData = Status{
	Playlist: 17,
	State:    "play",
	Song:     0,
	Songid:   2,
	Elapsed:  2507.560,
	Duration: 600,
	Bitrate:  128,
	Audio:    "44100:16:2",
	Error:    "",
}

func (s Stats) String() string {
	return fmt.Sprintf("artists: %d\nalbums: %d\ndb_playtime: %d\ndb_update: %d\nplaytime: %d\nsongs: %d\nuptime: %d\n",
		s.Artists, s.Albums, s.DBplaytime, s.DBupdate, s.Playtime, s.Songs, s.Uptime)
}

var statsTestData = Stats{
	Artists:    100,
	Albums:     110,
	DBplaytime: 380000,
	DBupdate:   1550336499,
	Playtime:   400,
	Songs:      1500,
	Uptime:     500,
}

func mockMPDServer(t *testing.T, server net.Conn) {
	fmt.Fprintln(server, "OK MPD 0.19.0")
	go mockMPDHandleConnection(t, server)
}

func mockMPDHandleConnection(t *testing.T, conn net.Conn) {
	var simpleCommands = regexp.MustCompile(`^(clearerror|ping|next|previous|stop)\r\n`)
	var pauseCommand = regexp.MustCompile(`^pause\s+(0|1)\r\n`)
	var playCommand = regexp.MustCompile(`^(play|playid)\s+(-1|\d+)\r\n`)
	var loadCommand = regexp.MustCompile(`^load\s+(\w+)\r\n`)
	var statusCommand = regexp.MustCompile(`^status\r\n`)
	var statsCommand = regexp.MustCompile(`^stats\r\n`)
	var currentSongCommand = regexp.MustCompile(`^currentsong\r\n`)

	buf := make([]byte, 4096)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}

		b := bytes.Trim(buf, "\x00")
		//fmt.Println(string(b))

		switch {
		case currentSongCommand.Match(b):
			fmt.Fprintf(conn, "%sOK\n", currentSongTestData)
		case statusCommand.Match(b):
			fmt.Fprintf(conn, "%sOK\n", statusTestData)
		case statsCommand.Match(b):
			fmt.Fprintf(conn, "%sOK\n", statsTestData)
		case simpleCommands.Match(b), pauseCommand.Match(b), playCommand.Match(b):
			fmt.Fprintln(conn, "OK")
		case loadCommand.Match(b):
			m := loadCommand.FindStringSubmatch(string(b))
			if m[1] == "nonexistent" {
				fmt.Fprintln(conn, "ACK [50@0] {load} No such playlist")
			} else {
				fmt.Fprintln(conn, "OK")
			}
		default:
			fmt.Fprintf(conn, "ACK {%s}", string(b))
		}
	}
	conn.Close()
}

func TestFailingConnection(t *testing.T) {
	_, err := Dial("localhost:1")
	assert.Error(t, err)
	_, err = Dial("host.notexists:6600")
	assert.Error(t, err)
}

func TestCurrentSongCommand(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	song, err := mpc.CurrentSong()
	if assert.NoError(t, err) {
		assert.Equal(t, currentSongTestData.File, song.File, "field CurrentSong.File")
		assert.Equal(t, currentSongTestData.ID, song.ID, "field CurrentSong.ID")
		assert.Equal(t, currentSongTestData.Name, song.Name, "field CurrentSong.Name")
		assert.Equal(t, currentSongTestData.Pos, song.Pos, "field CurrentSong.Pos")
		assert.Equal(t, currentSongTestData.Time, song.Time, "field CurrentSong.Time")
		assert.Equal(t, currentSongTestData.Title, song.Title, "field CurrentSong.Title")
	}
}

func TestStatusCommands(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	status, err := mpc.Status()
	if assert.NoError(t, err) {
		assert.Equal(t, statusTestData.Audio, status.Audio, "field Status.Audio")
		assert.Equal(t, statusTestData.State, status.State, "field Status.State")
		assert.Equal(t, statusTestData.Duration, status.Duration, "field Status.Duration")
		assert.Equal(t, statusTestData.Elapsed, status.Elapsed, "field Status.Elapsed")
		assert.Equal(t, statusTestData.Error, status.Error, "field Status.Error")
		assert.Equal(t, statusTestData.Playlist, status.Playlist, "field Status.Playlist")
		assert.Equal(t, statusTestData.State, status.State, "field Status.State")
		assert.Equal(t, statusTestData.Song, status.Song, "field Status.Song")
		assert.Equal(t, statusTestData.Songid, status.Songid, "field Status.Songid")
	}

	err = mpc.ClearError()
	assert.NoError(t, err)
}

func TestStatsCommand(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	stats, err := mpc.Stats()
	if assert.NoError(t, err) {
		assert.Equal(t, statsTestData.Artists, stats.Artists, "field Stats.Artists")
		assert.Equal(t, statsTestData.Albums, stats.Albums, "field Stats.Albums")
		assert.Equal(t, statsTestData.DBplaytime, stats.DBplaytime, "field Stats.DBplaytime")
		assert.Equal(t, statsTestData.DBupdate, stats.DBupdate, "field Stats.DBupdate")
		assert.Equal(t, statsTestData.Playtime, stats.Playtime, "field Stats.Playtime")
		assert.Equal(t, statsTestData.Songs, stats.Songs, "field Stats.Songs")
		assert.Equal(t, statsTestData.Uptime, stats.Uptime, "field Stats.Uptime")
	}
}

func TestControllingPlaybackCommands(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = mpc.Next()
	assert.NoError(t, err)

	err = mpc.Pause(false)
	assert.NoError(t, err)
	err = mpc.Pause(true)
	assert.NoError(t, err)

	err = mpc.Play(-1)
	assert.NoError(t, err)
	err = mpc.Play(0)
	assert.NoError(t, err)
	err = mpc.Play(10)
	assert.NoError(t, err)

	err = mpc.PlayID(-1)
	assert.NoError(t, err)
	err = mpc.PlayID(0)
	assert.NoError(t, err)
	err = mpc.PlayID(10)
	assert.NoError(t, err)

	err = mpc.Previous()
	assert.NoError(t, err)

	err = mpc.Stop()
	assert.NoError(t, err)
}

func TestStoredPlaylistsCommands(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = mpc.Load("GreatestHits")
	assert.NoError(t, err)
	err = mpc.Load("nonexistent")
	assert.Error(t, err)
}

func TestConnectionSettingsCommands(t *testing.T) {
	server, client := net.Pipe()
	go mockMPDServer(t, server)

	mpc, err := NewClient("localhost:6600", client)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = mpc.Ping()
	assert.NoError(t, err)
	err = mpc.Close()
	assert.NoError(t, err)
}
