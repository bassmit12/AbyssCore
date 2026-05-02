'use client'

import { useQuery, useMutation } from '@apollo/client/react'
import { GET_INVENTORY, USE_ITEM } from '@/lib/queries'

interface Props {
  heroId: string
  onHeroUpdated?: () => void
}

const RARITY_COLORS: Record<string, string> = {
  common: '#9ca3af',
  uncommon: '#34d399',
  rare: '#60a5fa',
  epic: '#c084fc',
  legendary: '#fbbf24',
}

const TYPE_ICONS: Record<string, string> = {
  potion: '🧪',
  weapon: '⚔️',
  armor: '🛡️',
  relic: '💎',
  gold: '💰',
}

export default function InventoryPanel({ heroId, onHeroUpdated }: Props) {
  const { data, loading, error, refetch } = useQuery(GET_INVENTORY, {
    variables: { heroId },
    fetchPolicy: 'network-only',
  })

  const [useItem] = useMutation(USE_ITEM, {
    onCompleted: () => {
      refetch()
      onHeroUpdated?.()
    },
  })

  if (loading as unknown as boolean) return (
    <div style={{ color: '#9ca3af', padding: '8px', fontSize: '14px' }}>Loading inventory...</div>
  )
  if (error as unknown as boolean) return (
    <div style={{ color: '#ef4444', padding: '8px', fontSize: '12px' }}>Inventory unavailable</div>
  )

  const items: any[] = (data as any)?.heroInventory?.items ?? []

  if (items.length === 0) return (
    <div style={{ color: '#6b7280', padding: '8px', fontSize: '13px', fontStyle: 'italic' }}>
      No items
    </div>
  )

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
      {items.map((item: any) => (
        <div
          key={item.id}
          style={{
            background: 'rgba(255,255,255,0.04)',
            border: `1px solid ${RARITY_COLORS[item.rarity] ?? '#374151'}`,
            borderRadius: '6px',
            padding: '6px 10px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: '13px', color: RARITY_COLORS[item.rarity] ?? '#e5e7eb', display: 'flex', gap: '6px', alignItems: 'center' }}>
              <span>{TYPE_ICONS[item.type] ?? '📦'}</span>
              <span style={{ fontWeight: 600 }}>{item.name}</span>
            </div>
            <div style={{ fontSize: '11px', color: '#9ca3af', marginTop: '2px' }}>{item.description}</div>
          </div>
          {(item.type === 'potion' || item.type === 'gold') && (
            <button
              onClick={() => useItem({ variables: { heroId, itemId: item.id } })}
              style={{
                background: '#059669',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                padding: '4px 8px',
                fontSize: '11px',
                cursor: 'pointer',
                whiteSpace: 'nowrap',
              }}
            >
              Use
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
