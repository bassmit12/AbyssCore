"use client"
import { ApolloProvider } from "@apollo/client/react"
import { SessionProvider, useSession, signIn } from "next-auth/react"
import { useEffect } from "react"
import { apolloClient } from "@/lib/apollo"

// Watches the session and auto-redirects to Keycloak if the refresh token fails.
function SessionWatcher() {
  const { data: session } = useSession()
  useEffect(() => {
    if ((session as any)?.error === "RefreshAccessTokenError") {
      console.warn("[SessionWatcher] refresh token invalid — re-authenticating")
      signIn("keycloak")
    }
  }, [session])
  return null
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    // refetchInterval: re-check session every 4 minutes so the server-side
    // JWT callback runs and refreshes the token before it expires (5 min default).
    <SessionProvider refetchInterval={4 * 60} refetchOnWindowFocus>
      <ApolloProvider client={apolloClient}>
        <SessionWatcher />
        {children}
      </ApolloProvider>
    </SessionProvider>
  )
}
