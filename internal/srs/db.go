package srs

import "fmt"

func (s *State) ShowPath() {
    fmt.Println(s.DBPath)
}