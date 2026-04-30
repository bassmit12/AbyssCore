"use client"
import { useQuery, useMutation } from "@apollo/client"
import { GET_INVENTORY, USE_ITEM } from "@/lib/queries"

interface Props {
  heroId: string
  onUse: () => void
}

const rarityColor: Record<string, string> = {
  common: "text-gray-300",
  uncommon: "text-green-400",
  rare: "text-purple-400",
}

export default function InventoryPanel({ heroId, onUse }: Props) {
  const { data, loading, refetch } = useQuery(GET_INVENTORY, { variables: { heroId } })
  const [useItem] = useMutation(USE_ITEM, {
    onCompleted: () => { refetch(); onUse() },
  })

  if (loading) return <div className="text-gray-600 text-sm">Loading...</div>

  const items = data?.inventory?.items ?? []

  if (items.length === 0) {
    return <div className="text-gray-600 text-sm">Your pack is empty.</div>
  }

  return (
    <div className="space-y-2">
      {items.map((item: { id: string; name: string; type: string; value: number; rarity: string }) => (
        <div
          key={item.id}
          className="flex items-center justify-between border border-gray-700 rounded-lg px-3 py-2 bg-gray-800"
        >
          <div>
            <div className={`font-medium ${rarityColor[item.rarity] ?? "text-gray-300"}`}>
              {item.name}
            </div>
            <div className="text-xs text-gray-500 capitalize">
              {item.type} · {item.rarity} · +{item.value}
            </div>
          </div>
          {item.type === "potion" && (
            <button
              onClick={() => useItem({ variables: { heroId, itemId: item.id } })}
              className="text-xs px-2 py-1 bg-green-800 hover:bg-green-700 rounded transition"
            >
              Use
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
