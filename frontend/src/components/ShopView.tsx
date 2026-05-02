"use client"

import { useMutation, useQuery } from "@apollo/client/react"
import { BUY_CARD, GET_SHOP_CARDS } from "@/lib/queries"

interface CardDef {
  id: string
  name: string
  class: string
  type: string
  cost: number
  effect: string
  rarity: string
  description: string
}

interface ShopItem {
  cardDef: CardDef
  price: number
}

interface Props {
  heroId: string
  nodeId: string
  heroGold: number
  onLeave: (updatedGold?: number) => void
}

const rarityColor: Record<string, string> = {
  common: "#9ca3af",
  uncommon: "#34d399",
  rare: "#a78bfa",
  epic: "#f59e0b",
}

const typeIcon: Record<string, string> = {
  attack: "⚔️",
  skill: "🛡️",
  power: "✨",
}

export default function ShopView({ heroId, heroGold, onLeave }: Props) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, loading, error: queryError } = useQuery(GET_SHOP_CARDS, {
    variables: { heroId },
    fetchPolicy: "network-only",
  })
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const error = queryError as { message?: string } | undefined

  const [buyCard, { loading: buying }] = useMutation(BUY_CARD)
  const [gold, setGold] = React.useState(heroGold)

  const handleBuy = async (item: ShopItem) => {
    if (gold < item.price) return
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const res = await buyCard({
        variables: {
          heroId,
          cardDefId: item.cardDef.id,
          price: item.price,
        },
      }) as { data?: { buyCard?: { gold: number } } }
      const newGold = res.data?.buyCard?.gold ?? gold - item.price
      setGold(newGold)
    } catch (e: unknown) {
      alert((e as Error).message ?? "Purchase failed")
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 text-white p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-yellow-400">🏪 Merchant&apos;s Shop</h1>
          <p className="text-gray-400 mt-1 text-sm">Cards for sale — choose wisely</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="bg-yellow-900/40 border border-yellow-600/30 rounded-lg px-4 py-2">
            <span className="text-yellow-400 font-bold text-lg">💰 {gold} gold</span>
          </div>
          <button
            onClick={() => onLeave(gold)}
            className="bg-gray-700 hover:bg-gray-600 text-white px-5 py-2 rounded-lg font-semibold transition"
          >
            Leave Shop
          </button>
        </div>
      </div>

      {(loading as unknown as boolean) ? (
        <div className="flex items-center justify-center h-64 text-gray-400">
          Loading inventory…
        </div>
      ) : (queryError as unknown as boolean) ? (
        <div className="text-red-400 text-center py-12">Error loading shop</div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
          {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
          {(data as any).shopCards.items.map((item: ShopItem) => {
            const canAfford = gold >= item.price
            const color = rarityColor[item.cardDef.rarity] ?? "#9ca3af"
            return (
              <div
                key={item.cardDef.id}
                className={`relative flex flex-col bg-gray-900 rounded-xl border transition-all ${
                  canAfford
                    ? "border-gray-600 hover:border-yellow-500/60 hover:scale-105"
                    : "border-gray-800 opacity-60"
                }`}
                style={{ boxShadow: canAfford ? `0 0 16px ${color}22` : undefined }}
              >
                {/* Rarity stripe */}
                <div
                  className="h-1 rounded-t-xl"
                  style={{ background: color }}
                />
                <div className="p-4 flex flex-col flex-1">
                  {/* Cost circle */}
                  <div className="flex items-start justify-between mb-3">
                    <div
                      className="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold bg-blue-900 border border-blue-400"
                    >
                      {item.cardDef.cost}
                    </div>
                    <span className="text-lg">{typeIcon[item.cardDef.type] ?? "🃏"}</span>
                  </div>
                  <h3 className="font-bold text-white mb-1">{item.cardDef.name}</h3>
                  <p className="text-xs text-gray-400 capitalize mb-2">
                    {item.cardDef.rarity} {item.cardDef.type}
                  </p>
                  <p className="text-sm text-gray-300 flex-1 leading-snug">
                    {item.cardDef.description || item.cardDef.effect}
                  </p>
                </div>
                {/* Buy button */}
                <div className="px-4 pb-4">
                  <button
                    disabled={!canAfford || buying}
                    onClick={() => handleBuy(item)}
                    className={`w-full py-2 rounded-lg font-semibold text-sm transition ${
                      canAfford
                        ? "bg-yellow-600 hover:bg-yellow-500 text-black"
                        : "bg-gray-700 text-gray-500 cursor-not-allowed"
                    }`}
                  >
                    {canAfford ? `Buy for 💰 ${item.price}` : `💰 ${item.price} (not enough)`}
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

// React needs to be in scope for useState
import React from "react"
