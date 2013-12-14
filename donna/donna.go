package main

import (
        `fmt`
        `github.com/michaeldv/donna`
)

func Repl() {
        var game *donna.Game
        var position *donna.Position

        for command := ``; ; command = `` {
                fmt.Print(`Donna> `)
                fmt.Scanf(`%s`, &command)
                switch command {
                case ``:
                case `exit`, `quit`:
                        return
                case `help`:
                        fmt.Println(`help: not implemented yet.`)
                case `new`:
                        game = donna.NewGame().Setup(`Ra1,Nb1,Bc1,Qd1,Ke1,Bf1,Nf3,Rh1,a2,b2,c2,d2,e4,f2,g2,h2`,
                                                     `Ra8,Nc6,Bc8,Qd8,Ke8,Bf8,Ng8,Rh8,a7,b7,c7,d7,e5,f7,g7,h7`)
                        position = game.Start()
                        fmt.Printf(`%s`, position)
                case `play`:
                        fmt.Println(`play: not implemented yet.`)
                default:
                        if position != nil {
                                move := donna.NewMoveFromString(command)
                                if move != nil {
                                        position = position.MakeMove(move)
                                        fmt.Printf(`%s`, position)
                                        best := game.Think(3, position)
                                        fmt.Printf("Computer plays %s\n", best)
                                        // move = donna.NewMoveFromString(move)
                                        // position = position.MakeMove(move)
                                        // fmt.Printf("%s", position)
                                } else {
                                        fmt.Printf("%s appears to be an invalid move.\n")
                                }
                        } else {
                                fmt.Println(`Please start a new game first.`)
                        }
                }
        }
}

func main() {
        donna.Settings.Log = false//true
        donna.Settings.Fancy = true

        // donna.NewGame().Setup(`Ka7,Qb1,Bg2`, `Ka5,b3,g3`).Think(4, nil) // Qb2
        // donna.NewGame().Setup(`Kh5,Qg7,Be5,f2,f3`, `Kh1`).Think(4, nil) // Bh2
        // donna.NewGame().Setup(`Kd3,Rd8,a5,b2,f2,g5`, `Kd1`).Think(4, nil) // Rd4
        Repl()
}
