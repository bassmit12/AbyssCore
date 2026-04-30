import { gql } from "@apollo/client"

export const CREATE_HERO = gql`
  mutation CreateHero($name: String!, $class: HeroClass!) {
    createHero(name: $name, class: $class) {
      id name class hp maxHp level xp
    }
  }
`

export const GET_HERO = gql`
  query GetHero($id: ID!) {
    hero(id: $id) {
      id name class hp maxHp level xp x y alive dungeonId
    }
  }
`

export const START_DUNGEON = gql`
  mutation StartDungeon($heroId: ID!) {
    startDungeon(heroId: $heroId) {
      dungeonId level rooms {
        x y hasChest exits monsters
      }
    }
  }
`

export const MOVE_HERO = gql`
  mutation MoveHero($heroId: ID!, $direction: String!) {
    moveHero(heroId: $heroId, direction: $direction) {
      id x y hp maxHp level xp
    }
  }
`

export const ATTACK = gql`
  mutation Attack($heroId: ID!, $monsterId: ID!) {
    attack(heroId: $heroId, monsterId: $monsterId) {
      heroDamageDealt monsterDamageBack monsterDied heroDied message
    }
  }
`

export const GET_FLOOR = gql`
  query GetFloor($dungeonId: ID!, $level: Int!) {
    dungeonFloor(dungeonId: $dungeonId, level: $level) {
      level rooms { x y hasChest exits monsters }
    }
  }
`

export const GET_INVENTORY = gql`
  query GetInventory($heroId: ID!) {
    inventory(heroId: $heroId) {
      items { id name type value rarity }
    }
  }
`

export const USE_ITEM = gql`
  mutation UseItem($heroId: ID!, $itemId: ID!) {
    useItem(heroId: $heroId, itemId: $itemId)
  }
`

export const GET_LEADERBOARD = gql`
  query GetLeaderboard {
    leaderboard {
      id heroName playerName floorsCleared monstersKilled itemsFound score
    }
  }
`

export const COMBAT_EVENTS = gql`
  subscription CombatEvents($heroId: ID!) {
    combatEvent(heroId: $heroId) {
      heroId monsterId heroDamageDealt monsterDamageBack monsterDied heroDied message
    }
  }
`
