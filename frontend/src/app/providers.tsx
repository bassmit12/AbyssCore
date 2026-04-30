"use client"
import { ApolloProvider } from "@apollo/client"
import { SessionProvider } from "next-auth/react"
import { apolloClient } from "@/lib/apollo"

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <ApolloProvider client={apolloClient}>
        {children}
      </ApolloProvider>
    </SessionProvider>
  )
}
