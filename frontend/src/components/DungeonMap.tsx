"use client"

interface Room {
  x: number
  y: number
  hasChest: boolean
  exits: string[]
  monsters: string[]
}

interface Floor {
  level: number
  rooms: Room[][]
}

interface Hero {
  x: number
  y: number
  hp: number
  maxHp: number
  alive: boolean
}

interface Props {
  floor?: Floor
  hero?: Hero
  onAttack: (monsterId: string) => void
}

const CELL = 48

export default function DungeonMap({ floor, hero, onAttack }: Props) {
  if (!floor) {
    return <div className="text-gray-600 text-sm">Generating floor...</div>
  }

  const rows = floor.rooms
  const height = rows.length
  const width = rows[0]?.length ?? 0

  return (
    <div className="relative border border-gray-800 rounded-lg overflow-auto bg-gray-900">
      <div
        style={{ width: width * CELL, height: height * CELL, position: "relative" }}
      >
        {rows.map((row, y) =>
          row.map((room, x) => {
            if (!room || (!room.exits && !room.hasChest && !room.monsters?.length)) {
              // Empty cell (wall)
              return (
                <div
                  key={`${x}-${y}`}
                  style={{
                    position: "absolute",
                    left: x * CELL,
                    top: y * CELL,
                    width: CELL,
                    height: CELL,
                  }}
                  className="bg-gray-950"
                />
              )
            }

            const isHere = hero && hero.x === x && hero.y === y
            const hasMonster = room.monsters && room.monsters.length > 0

            return (
              <div
                key={`${x}-${y}`}
                style={{
                  position: "absolute",
                  left: x * CELL,
                  top: y * CELL,
                  width: CELL,
                  height: CELL,
                }}
                className={`border border-gray-700 flex items-center justify-center text-lg cursor-default
                  ${hasMonster ? "bg-red-950 hover:bg-red-900 cursor-pointer" : "bg-gray-800"}
                `}
                onClick={() => {
                  if (hasMonster && isHere) {
                    onAttack(room.monsters[0])
                  }
                }}
                title={
                  hasMonster
                    ? isHere
                      ? "Click to attack!"
                      : "Monster lurks here"
                    : room.hasChest
                    ? "Chest"
                    : "Room"
                }
              >
                {isHere ? (
                  <span className="text-purple-400 drop-shadow-lg">⚔</span>
                ) : hasMonster ? (
                  <span className="text-red-400">☠</span>
                ) : room.hasChest ? (
                  <span className="text-yellow-400">⬡</span>
                ) : (
                  <span className="text-gray-700 text-xs">{room.exits?.includes("south") ? "↓" : ""}</span>
                )}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
