'use client'

import { useState } from 'react'
import { useQuery, useMutation } from '@apollo/client/react'
import { GET_RANDOM_EVENT, RESOLVE_EVENT } from '@/lib/queries'

interface Props {
  heroId: string
  onDone: (outcome: { goldDelta: number; hpDelta: number }) => void
}

export default function EventView({ heroId, onDone }: Props) {
  const [resolved, setResolved] = useState<any>(null)
  const [resolving, setResolving] = useState(false)

  const { data, loading, error } = useQuery(GET_RANDOM_EVENT, {
    fetchPolicy: 'network-only',
  })

  const [resolveEvent] = useMutation(RESOLVE_EVENT, {
    onCompleted: (data) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setResolved((data as any).resolveEvent)
      setResolving(false)
    },
  })

  if (loading as unknown as boolean) return (
    <div style={containerStyle}>
      <div style={{ color: '#9ca3af' }}>Generating event...</div>
    </div>
  )
  if (error as unknown as boolean) return (
    <div style={containerStyle}>
      <div style={{ color: '#ef4444' }}>Event failed to load.</div>
      <button style={btnStyle('#6b7280')} onClick={() => onDone({ goldDelta: 0, hpDelta: 0 })}>Leave</button>
    </div>
  )

  const event = (data as any)?.randomEvent

  if (resolved) {
    return (
      <div style={containerStyle}>
        <div style={{ fontSize: '28px', marginBottom: '12px' }}>📜</div>
        <h2 style={{ color: '#fcd34d', marginBottom: '12px' }}>Outcome</h2>
        <p style={{ color: '#e5e7eb', marginBottom: '16px', maxWidth: '400px', textAlign: 'center', lineHeight: 1.6 }}>
          {resolved.description}
        </p>
        <div style={{ display: 'flex', gap: '24px', marginBottom: '24px' }}>
          {resolved.goldDelta !== 0 && (
            <span style={{ color: resolved.goldDelta > 0 ? '#fbbf24' : '#ef4444', fontSize: '16px' }}>
              {resolved.goldDelta > 0 ? '+' : ''}{resolved.goldDelta} 💰
            </span>
          )}
          {resolved.hpDelta !== 0 && (
            <span style={{ color: resolved.hpDelta > 0 ? '#34d399' : '#ef4444', fontSize: '16px' }}>
              {resolved.hpDelta > 0 ? '+' : ''}{resolved.hpDelta} ❤️
            </span>
          )}
          {resolved.goldDelta === 0 && resolved.hpDelta === 0 && (
            <span style={{ color: '#9ca3af' }}>No change</span>
          )}
        </div>
        <button
          style={btnStyle('#6d28d9')}
          onClick={() => onDone({ goldDelta: resolved.goldDelta, hpDelta: resolved.hpDelta })}
        >
          Continue
        </button>
      </div>
    )
  }

  return (
    <div style={containerStyle}>
      <div style={{ fontSize: '32px', marginBottom: '12px' }}>📜</div>
      <h2 style={{ color: '#fcd34d', marginBottom: '12px', fontSize: '22px' }}>
        {event?.title ?? 'Strange Encounter'}
      </h2>
      <p style={{ color: '#d1d5db', marginBottom: '24px', maxWidth: '420px', textAlign: 'center', lineHeight: 1.7 }}>
        {event?.description}
      </p>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', width: '100%', maxWidth: '380px' }}>
        {event?.choices?.map((choice: any) => (
          <button
            key={choice.id}
            disabled={resolving}
            onClick={() => {
              setResolving(true)
              resolveEvent({ variables: { heroId, eventId: event.id, choiceId: choice.id } })
            }}
            style={{
              ...choiceBtnStyle,
              opacity: resolving ? 0.5 : 1,
              cursor: resolving ? 'not-allowed' : 'pointer',
            }}
          >
            <span style={{ fontWeight: 600 }}>{choice.label}</span>
            {choice.description && (
              <span style={{ fontSize: '11px', color: '#9ca3af', marginTop: '3px' }}>{choice.description}</span>
            )}
          </button>
        ))}
        <button
          style={btnStyle('#374151')}
          disabled={resolving}
          onClick={() => onDone({ goldDelta: 0, hpDelta: 0 })}
        >
          Leave (no effect)
        </button>
      </div>
    </div>
  )
}

const containerStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  minHeight: '60vh',
  padding: '24px',
}

const choiceBtnStyle: React.CSSProperties = {
  background: 'rgba(109,40,217,0.2)',
  border: '1px solid #6d28d9',
  borderRadius: '8px',
  color: '#e5e7eb',
  padding: '12px 16px',
  cursor: 'pointer',
  display: 'flex',
  flexDirection: 'column',
  textAlign: 'left',
  transition: 'background 0.15s',
  fontSize: '14px',
}

function btnStyle(bg: string): React.CSSProperties {
  return {
    background: bg,
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    padding: '10px 20px',
    cursor: 'pointer',
    fontSize: '14px',
  }
}
