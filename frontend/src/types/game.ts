"use client"

// Types shared across combat components
export interface Card {
  id: string
  defId: string
  name: string
  cost: number
  type: string
  effect: string
}

export interface CardDefinition {
  id: string
  name: string
  class: string
  type: string
  cost: number
  effect: string
  rarity: string
  description: string
}

export interface StatusEffect {
  name: string
  stacks: number
}

export interface MonsterIntent {
  type: string
  value: number
}

export interface EncounterMonster {
  id: string
  name: string
  hp: number
  maxHp: number
  block: number
  status: string
  intents: MonsterIntent[]
}

export interface HeroCombatState {
  hp: number
  maxHp: number
  block: number
  energy: number
  maxEnergy: number
  hand: Card[]
  drawPileCount: number
  discardPileCount: number
  statuses: StatusEffect[]
}

export interface EncounterState {
  encounterId: string
  heroState: HeroCombatState
  monsters: EncounterMonster[]
  turnNumber: number
  status: string
  message: string
}

export interface MapNode {
  id: string
  floor: number
  position: number
  type: string
  visited: boolean
  available: boolean
}

export interface MapEdge {
  fromNodeId: string
  toNodeId: string
}

export interface FloorGraph {
  runId: string
  heroId: string
  currentNodeId: string | null
  nodes: MapNode[]
  edges: MapEdge[]
}
