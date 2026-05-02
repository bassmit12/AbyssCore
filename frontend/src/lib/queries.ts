import { gql } from "@apollo/client"

// ─── Hero ─────────────────────────────────────────────────────────────────────

export const CREATE_HERO = gql`
  mutation CreateHero($name: String!, $class: HeroClass!) {
    createHero(name: $name, class: $class) {
      id name class hp maxHp level xp gold alive
    }
  }
`

export const GET_HERO = gql`
  query GetHero($id: ID!) {
    hero(id: $id) {
      id name class hp maxHp level xp gold alive runId
    }
  }
`

// ─── Map / Run ────────────────────────────────────────────────────────────────

export const START_RUN = gql`
  mutation StartRun($heroId: ID!) {
    startRun(heroId: $heroId) {
      runId heroId currentNodeId
      nodes { id floor position type visited available }
      edges { fromNodeId toNodeId }
    }
  }
`

export const TRAVEL_TO_NODE = gql`
  mutation TravelToNode($heroId: ID!, $nodeId: ID!) {
    travelToNode(heroId: $heroId, nodeId: $nodeId) {
      runId heroId currentNodeId
      nodes { id floor position type visited available }
      edges { fromNodeId toNodeId }
    }
  }
`

export const GET_FLOOR_GRAPH = gql`
  query GetFloorGraph($heroId: ID!) {
    floorGraph(heroId: $heroId) {
      runId heroId currentNodeId
      nodes { id floor position type visited available }
      edges { fromNodeId toNodeId }
    }
  }
`

// ─── Card Combat ─────────────────────────────────────────────────────────────

const ENCOUNTER_STATE_FRAGMENT = gql`
  fragment EncounterStateFields on EncounterState {
    encounterId
    heroState {
      hp maxHp block energy maxEnergy
      drawPileCount discardPileCount
      hand { id defId name cost type effect }
      statuses { name stacks }
    }
    monsters {
      id name hp maxHp block status
      intents { type value }
    }
    turnNumber status message
  }
`

export const START_ENCOUNTER = gql`
  ${ENCOUNTER_STATE_FRAGMENT}
  mutation StartEncounter($heroId: ID!, $nodeId: ID!) {
    startEncounter(heroId: $heroId, nodeId: $nodeId) {
      ...EncounterStateFields
    }
  }
`

export const PLAY_CARD = gql`
  ${ENCOUNTER_STATE_FRAGMENT}
  mutation PlayCard($encounterId: ID!, $heroId: ID!, $cardId: ID!, $targetId: ID) {
    playCard(encounterId: $encounterId, heroId: $heroId, cardId: $cardId, targetId: $targetId) {
      ...EncounterStateFields
    }
  }
`

export const END_TURN = gql`
  ${ENCOUNTER_STATE_FRAGMENT}
  mutation EndTurn($encounterId: ID!, $heroId: ID!) {
    endTurn(encounterId: $encounterId, heroId: $heroId) {
      ...EncounterStateFields
    }
  }
`

export const GET_ENCOUNTER_STATE = gql`
  ${ENCOUNTER_STATE_FRAGMENT}
  query GetEncounterState($encounterId: ID!) {
    encounterState(encounterId: $encounterId) {
      ...EncounterStateFields
    }
  }
`

// ─── Card Rewards ─────────────────────────────────────────────────────────────

export const GET_CARD_REWARDS = gql`
  query GetCardRewards($heroId: ID!) {
    cardRewards(heroId: $heroId) {
      cards { id name class type cost effect rarity description }
    }
  }
`

export const PICK_CARD_REWARD = gql`
  mutation PickCardReward($encounterId: ID!, $heroId: ID!, $cardDefId: ID!) {
    pickCardReward(encounterId: $encounterId, heroId: $heroId, cardDefId: $cardDefId) {
      heroId
      cards { id defId name cost type effect }
    }
  }
`

export const SKIP_CARD_REWARD = gql`
  mutation SkipCardReward($encounterId: ID!, $heroId: ID!) {
    skipCardReward(encounterId: $encounterId, heroId: $heroId)
  }
`

// ─── Deck & Relics ────────────────────────────────────────────────────────────

export const GET_HERO_DECK = gql`
  query GetHeroDeck($heroId: ID!) {
    heroDeck(heroId: $heroId) {
      heroId
      cards { id defId name cost type effect }
    }
  }
`

export const GET_HERO_RELICS = gql`
  query GetHeroRelics($heroId: ID!) {
    heroRelics(heroId: $heroId) {
      id defId name rarity description
    }
  }
`

// ─── Shop ─────────────────────────────────────────────────────────────────────

export const GET_SHOP_CARDS = gql`
  query GetShopCards($heroId: ID!) {
    shopCards(heroId: $heroId) {
      items {
        cardDef { id name class type cost effect rarity description }
        price
      }
    }
  }
`

export const GET_INVENTORY = gql`
  query GetInventory($heroId: ID!) {
    heroInventory(heroId: $heroId) {
      heroId
      items {
        id
        name
        type
        value
        rarity
        description
      }
    }
  }
`;

export const USE_ITEM = gql`
  mutation UseItem($heroId: ID!, $itemId: ID!) {
    useItem(heroId: $heroId, itemId: $itemId)
  }
`;

export const GET_RANDOM_EVENT = gql`
  query GetRandomEvent {
    randomEvent {
      id
      title
      description
      choices {
        id
        label
        description
      }
    }
  }
`;

export const RESOLVE_EVENT = gql`
  mutation ResolveEvent($heroId: ID!, $eventId: ID!, $choiceId: ID!) {
    resolveEvent(heroId: $heroId, eventId: $eventId, choiceId: $choiceId) {
      description
      goldDelta
      hpDelta
    }
  }
`;
export const BUY_CARD = gql`
  mutation BuyCard($heroId: ID!, $cardDefId: ID!, $price: Int!) {
    buyCard(heroId: $heroId, cardDefId: $cardDefId, price: $price) {
      id name class hp maxHp level xp gold alive runId
    }
  }
`

// ─── Submit Score ─────────────────────────────────────────────────────────────

export const SUBMIT_SCORE = gql`
  mutation SubmitScore($heroId: ID!) {
    submitScore(heroId: $heroId) {
      id heroName floorsCleared monstersKilled score
    }
  }
`

// ─── Leaderboard ─────────────────────────────────────────────────────────────

export const GET_LEADERBOARD = gql`
  query GetLeaderboard {
    leaderboard {
      id heroName playerName floorsCleared monstersKilled score
    }
  }
`

// ─── Subscriptions ───────────────────────────────────────────────────────────

export const ENCOUNTER_UPDATED = gql`
  ${ENCOUNTER_STATE_FRAGMENT}
  subscription EncounterUpdated($encounterId: ID!) {
    encounterUpdated(encounterId: $encounterId) {
      ...EncounterStateFields
    }
  }
`
