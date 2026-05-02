"use client"
import { ApolloClient, InMemoryCache, HttpLink, split } from "@apollo/client"
import { setContext } from "@apollo/client/link/context"
import { onError } from "@apollo/client/link/error"
import { GraphQLWsLink } from "@apollo/client/link/subscriptions"
import { createClient } from "graphql-ws"
import { getMainDefinition } from "@apollo/client/utilities"
import { getSession, signIn } from "next-auth/react"

const httpLink = new HttpLink({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_URL ?? "http://localhost:4001/graphql",
})

// Auto-refresh: if the session has a RefreshAccessTokenError, silently re-trigger
// the Keycloak login flow so the user never has to log out manually.
const errorLink = onError((errorResponse) => {
  const graphQLErrors = (errorResponse as any).graphQLErrors
  if (graphQLErrors) {
    for (const err of graphQLErrors) {
      if (err?.extensions?.code === "UNAUTHENTICATED") {
        signIn("keycloak")
      }
    }
  }
})

const authLink = setContext(async (_, { headers }) => {
  const session = await getSession()

  // If the server-side refresh failed, trigger re-login
  if ((session as any)?.error === "RefreshAccessTokenError") {
    console.warn("[apollo] refresh token expired — re-authenticating")
    signIn("keycloak")
    return { headers }
  }

  const token = (session as any)?.accessToken
  return {
    headers: {
      ...headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  }
})

const wsLink =
  typeof window !== "undefined"
    ? new GraphQLWsLink(
        createClient({
          url: (process.env.NEXT_PUBLIC_GRAPHQL_WS_URL ?? "ws://localhost:4001/graphql"),
          connectionParams: async () => {
            const session = await getSession()
            const token = (session as any)?.accessToken
            return token ? { Authorization: `Bearer ${token}` } : {}
          },
        })
      )
    : null

const splitLink =
  wsLink
    ? split(
        ({ query }) => {
          const def = getMainDefinition(query)
          return def.kind === "OperationDefinition" && def.operation === "subscription"
        },
        wsLink,
        errorLink.concat(authLink.concat(httpLink))
      )
    : errorLink.concat(authLink.concat(httpLink))

export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
})
