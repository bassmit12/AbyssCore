import { auth } from "@/auth"
import { NextResponse } from "next/server"

export default auth((req) => {
  const isLoggedIn = !!req.auth
  const isAuthRoute = req.nextUrl.pathname.startsWith("/api/auth")
  const isPublicRoute =
    req.nextUrl.pathname === "/" ||
    req.nextUrl.pathname === "/leaderboard" ||
    isAuthRoute

  if (!isLoggedIn && !isPublicRoute) {
    return NextResponse.redirect(new URL("/api/auth/signin", req.nextUrl))
  }

  return NextResponse.next()
})

export const config = {
  // Protect all routes except static files and Next internals
  matcher: ["/((?!_next/static|_next/image|favicon.ico|public/).*)"],
}
