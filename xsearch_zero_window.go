// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import(`fmt`)

// Search with zero window.
func (p *Position) xSearchWithZeroWindow(beta, depth int) int {
        if p.isRepetition() {
                return 0
        }

        bestScore := Ply() - Checkmate
        if bestScore >= beta {
                return bestScore
        }

        moveCount := 0
        gen := p.StartMoveGen(Ply()).GenerateMoves().rank()
        for move := gen.NextMove(); move != 0; move = gen.NextMove() {
                if position := p.MakeMove(move); position != nil {
                        fmt.Printf("%*szero/%s> depth: %d, ply: %d, move: %s\n", Ply()*2, ` `, C(p.color), depth, Ply(), move)
                        inCheck := position.isInCheck(position.color)
                        reducedDepth := depth - 1
                        if inCheck {
                                reducedDepth++
                        }

                        moveScore := 0
                        if reducedDepth == 0 {
                                moveScore = -position.xSearchQuiescence(-beta, 1 - beta, true)
                        } else if inCheck {
                                moveScore = -position.xSearchInCheck(1 - beta, reducedDepth)
                        } else {
                                moveScore = -position.xSearchWithZeroWindow(1 - beta, reducedDepth)
                        }
                        position.TakeBack(move)
                        moveCount++

                        if moveScore > bestScore {
                                if moveScore >= beta {
                                        return moveScore
                                }
                                bestScore = moveScore
                        }
                }
        } // next move.

        if moveCount == 0 {
                return 0
        }

        return bestScore
}
