// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import ()

type PawnEntry struct {
	hash     uint64 	// Pawn hash key.
	score    Score 		// Static score for the given pawn structure.
	passers  [2]Bitmask 	// Passed pawn bitmasks for both sides.
}

var pawnCache [8192]PawnEntry

func (e *Evaluation) analyzePawns() {
	key := e.position.hashPawn

	// Since pawn hash is fairly small we can use much faster 32-bit index.
	index := uint32(key) % uint32(len(pawnCache))
	e.pawns = &pawnCache[index]

	// Bypass pawns cache if evaluation tracing is enabled.
	if e.pawns.hash != key || Settings.Trace {
		white, black := e.pawnStructure(White), e.pawnStructure(Black)
		e.pawns.score.clear().add(white).subtract(black)
		e.pawns.hash = key
		if Settings.Trace {
			e.checkpoint(`Pawns`, Total{white, black})
		}
	}

	e.score.add(e.pawns.score)
}

func (e *Evaluation) analyzePassers() {
	var white, black Score

	if Settings.Trace {
		defer func() {
			e.checkpoint(`Passers`, Total{white, black})
		}()
	}

	white, black = e.pawnPassers(White), e.pawnPassers(Black)

	e.score.add(white).subtract(black)
}

// Calculates extra bonus and penalty based on pawn structure. Specifically,
// a bonus is awarded for passed pawns, and penalty applied for isolated and
// doubled pawns.
func (e *Evaluation) pawnStructure(color int) (score Score) {
	hisPawns := e.position.outposts[pawn(color)]
	herPawns := e.position.outposts[pawn(color^1)]
	e.pawns.passers[color] = 0

	pawns := hisPawns
	for pawns != 0 {
		square := pawns.pop()
		row, col := Coordinate(square)

		// Penalty if the pawn is isolated, i.e. has no friendly pawns
		// on adjacent files. The penalty goes up if isolated pawn is
		// exposed on semi-open file.
		isolated := (maskIsolated[col] & hisPawns == 0)
		exposed := (maskInFront[color][square] & herPawns == 0)
		if isolated {
			if !exposed {
				score.subtract(penaltyIsolatedPawn[col])
			} else {
				score.subtract(penaltyWeakIsolatedPawn[col])
			}
		}

		// Penalty if the pawn is doubled, i.e. there is another friendly
		// pawn in front of us. The penalty goes up if doubled pawns are
		// isolated.
		doubled := (maskInFront[color][square] & hisPawns != 0)
		if doubled {
			score.subtract(penaltyDoubledPawn[col])
		}

		// Bonus if the pawn is supported by friendly pawn(s) on the same
		// or previous ranks.
		supported := (maskIsolated[col] & (maskRank[row] | maskRank[row].pushed(color^1)) & hisPawns != 0)
		if supported {
			flip := Flip(color, square)
			score.add(Score{bonusSupportedPawn[flip], bonusSupportedPawn[flip]})
		}

		// The pawn is passed if a) there are no enemy pawns in the same
		// and adjacent columns; and b) there are no same color pawns in
		// front of us.
		passed := (maskPassed[color][square] & herPawns == 0 && !doubled)
		if passed {
			e.pawns.passers[color] |= bit[square]
		}

		// Penalty if the pawn is backward.
		if (!passed && !supported && !isolated) {

			// Backward pawn should not be attacking enemy pawns.
			if pawnMoves[color][square] & herPawns == 0 {

				// Backward pawn should not have friendly pawns behind.
				if maskPassed[color^1][square] & maskIsolated[col] & hisPawns == 0 {

					// Backward pawn should face enemy pawns on the next two ranks
					// preventing its advance.
					enemy := pawnMoves[color][square].pushed(color)
					if (enemy | enemy.pushed(color)) & herPawns != 0 {
						if !exposed {
							score.subtract(penaltyBackwardPawn[col])
						} else {
							score.subtract(penaltyWeakBackwardPawn[col])
						}
					}
				}
			}
		}

		// TODO: Bonus if the pawn has good chance to become a passed pawn.
	}

	return
}

func (e *Evaluation) pawnPassers(color int) (score Score) {
	p := e.position

	pawns := e.pawns.passers[color]
	for pawns != 0 {
		square := pawns.pop()
		row := RelRow(square, color)
		bonus := bonusPassedPawn[row]

		if row > A2H2 {
			nextSquare := square + eight[color]

			// Check if the pawn can step forward.
			if p.board.isClear(nextSquare) {

				// Assume all squares in front of the pawn are under attack.
				attacked := maskInFront[color][square]
				protected := attacked & e.attacks[color]

				// Check if the assumption is true and whether there is a queen
				// or a rook attacking our passed pawn from behind.
				enemy := maskInFront[color^1][square] & (p.outposts[queen(color^1)] | p.outposts[rook(color^1)])
				if enemy == 0 || enemy & p.rookMoves(square) == 0 {

					// Since nobody attacks the pawn from behind adjust the attacked
					// bitmask to only include squares attacked or occupied by the enemy.
					attacked &= (e.attacks[color^1] | p.outposts[color^1])

				}

				// Boost the bonus if passed pawn is free to run to the 8th rank
				// or at least safely step forward.
				extra := 0
				if attacked == 0 {
					extra = 10
				} else if attacked.isClear(nextSquare) {
					extra = 6
				}

				// Boost the bonus even more if all the squares in front of the
				// pawn are protected. If not, see if next square is protected.
				if protected == maskInFront[color][square] {
					extra += 4
				} else if protected.isSet(nextSquare) {
					extra += 2
				}
				if extra > 0 {
					bonus.adjust(extra * extraPassedPawn[row])
				}
			}
			// TODO: Adjust bonus based on proximity of both kings.
		}
		score.add(bonus)
	}

	// Penalty for blocked pawns.
	blocked := (p.outposts[pawn(color)].pushed(color) & p.board).count()
	score.subtract(pawnBlocked.times(blocked))

	return
}

