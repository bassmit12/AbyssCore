import NextAuth from "next-auth"

async function refreshAccessToken(token: any) {
  try {
    const issuer = process.env.KEYCLOAK_ISSUER!
    const tokenUrl = `${issuer}/protocol/openid-connect/token`

    const response = await fetch(tokenUrl, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "refresh_token",
        client_id: process.env.KEYCLOAK_CLIENT_ID!,
        client_secret: process.env.KEYCLOAK_CLIENT_SECRET ?? "",
        refresh_token: token.refreshToken,
      }),
    })

    const refreshed = await response.json()

    if (!response.ok) {
      console.error("[auth] token refresh failed:", refreshed)
      return { ...token, error: "RefreshAccessTokenError" }
    }

    console.log("[auth] token refreshed successfully")
    return {
      ...token,
      accessToken: refreshed.access_token,
      refreshToken: refreshed.refresh_token ?? token.refreshToken,
      accessTokenExpires: Date.now() + (refreshed.expires_in - 30) * 1000,
      error: undefined,
    }
  } catch (error) {
    console.error("[auth] refresh error:", error)
    return { ...token, error: "RefreshAccessTokenError" }
  }
}

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    {
      id: "keycloak",
      name: "Keycloak",
      type: "oidc",
      clientId: process.env.KEYCLOAK_CLIENT_ID!,
      clientSecret: process.env.KEYCLOAK_CLIENT_SECRET ?? "",
      issuer: process.env.KEYCLOAK_ISSUER,
      wellKnown: process.env.KEYCLOAK_WELL_KNOWN || `${process.env.KEYCLOAK_ISSUER}/.well-known/openid-configuration`,
      style: { logo: "/keycloak.svg", bg: "#fff", text: "#000" },
    },
  ],
  callbacks: {
    async jwt({ token, account }) {
      // First login — store all tokens
      if (account) {
        return {
          ...token,
          accessToken: account.access_token,
          refreshToken: account.refresh_token,
          // expires_at is seconds since epoch; subtract 30s buffer
          accessTokenExpires: account.expires_at
            ? account.expires_at * 1000 - 30_000
            : Date.now() + 4 * 60 * 1000,
        }
      }

      // No refresh token stored (old session pre-upgrade) — force re-login
      if (!token.refreshToken) {
        return { ...token, error: "RefreshAccessTokenError" }
      }

      // Token still valid
      if (Date.now() < (token.accessTokenExpires as number)) {
        return token
      }

      // Token expired — refresh
      console.log("[auth] access token expired, refreshing...")
      return refreshAccessToken(token)
    },

    async session({ session, token }) {
      session.accessToken = token.accessToken as string
      if (token.error) {
        ;(session as any).error = token.error
      }
      return session
    },
  },
})
