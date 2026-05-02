"use client"
import { useQuery, useMutation } from "@apollo/client/react"
import { GET_CARD_REWARDS, PICK_CARD_REWARD, SKIP_CARD_REWARD } from "@/lib/queries"
import type { CardDefinition } from "@/types/game"

const RARITY_COLOR: Record<string, string> = {
  COMMON:   "border-gray-600",
  UNCOMMON: "border-blue-500",
  RARE:     "border-yellow-500",
}

const TYPE_ICON: Record<string, string> = {
  ATTACK: "⚔️",
  SKILL:  "🔷",
  POWER:  "⚡",
}

interface Props {
  encounterId: string
  heroId: string
  onDone: () => void
}

export default function CardReward({ encounterId, heroId, onDone }: Props) {
  const { data, loading } = useQuery(GET_CARD_REWARDS, {
    variables: { heroId },
    fetchPolicy: "network-only",
  })

  const [pick, { loading: picking }] = useMutation(PICK_CARD_REWARD, {
    onCompleted: onDone,
  })

  const [skip, { loading: skipping }] = useMutation(SKIP_CARD_REWARD, {
    onCompleted: onDone,
  })

  if (loading) {
    return (
      <div className="flex items-center justify-center h-40 text-gray-400">
        Loading rewards...
      </div>
    )
  }

  const cards: CardDefinition[] = (data as { cardRewards?: { cards?: CardDefinition[] } } | undefined)?.cardRewards?.cards ?? []

  return (
    <div className="flex flex-col items-center gap-6 p-6">
      <h2 className="text-2xl font-bold text-yellow-400">Choose a Card</h2>

      <div className="flex gap-4 flex-wrap justify-center">
        {cards.map((card) => {
          const rarityClass = RARITY_COLOR[card.rarity] ?? RARITY_COLOR.COMMON
          return (
            <button
              key={card.id}
              disabled={picking || skipping}
              onClick={() => pick({ variables: { encounterId, heroId, cardDefId: card.id } })}
              className={`
                flex flex-col gap-2 p-4 rounded-xl border-2 w-44 text-left
                bg-gray-900 hover:bg-gray-800 transition-all cursor-pointer hover:-translate-y-1
                ${rarityClass}
                ${picking ? "opacity-50" : ""}
              `}
            >
              {/* Cost + type */}
              <div className="flex justify-between items-center">
                <span className="bg-blue-700 rounded-full w-6 h-6 flex items-center justify-center text-xs font-bold">
                  {card.cost}
                </span>
                <span title={card.type}>{TYPE_ICON[card.type] ?? "🃏"}</span>
              </div>

              {/* Name */}
              <div className="text-sm font-bold text-white">{card.name}</div>

              {/* Description */}
              <div className="text-xs text-gray-400 leading-tight">{card.description || card.effect}</div>

              {/* Rarity */}
              <div className="text-[10px] text-gray-500 uppercase tracking-wide mt-auto">
                {card.rarity.toLowerCase()} · {card.class.toLowerCase()}
              </div>
            </button>
          )
        })}
      </div>

      <button
        disabled={picking || skipping}
        onClick={() => skip({ variables: { encounterId, heroId } })}
        className="text-gray-500 hover:text-gray-300 text-sm underline underline-offset-2 transition"
      >
        Skip reward
      </button>
    </div>
  )
}
