import { useState } from "react";
import { Link, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { ExchangePage } from "./pages/ExchangePage";
import { MarketsPage } from "./pages/MarketsPage";
import { PortfolioPage } from "./pages/PortfolioPage";
import { StationDetailPage } from "./pages/StationDetailPage";
import { StationsPage } from "./pages/StationsPage";
import { LoginModal } from "./components/LoginModal";
import { useAuth } from "./lib/auth";
import { ChromeProvider } from "./lib/chrome";

function AppShell({ children }) {
  const location = useLocation();
  const { session, logout } = useAuth();
  const [loginOpen, setLoginOpen] = useState(false);
  const [authMode, setAuthMode] = useState("login");

  const activePortfolio = location.pathname === "/portfolio";
  const activeStations =
    location.pathname === "/" ||
    location.pathname === "/stations" ||
    location.pathname.startsWith("/station/");
  const activeMarkets = location.pathname === "/markets" || location.pathname.startsWith("/contracts/");

  const openLogin = () => {
    setAuthMode("login");
    setLoginOpen(true);
  };

  const openSignup = () => {
    setAuthMode("signup");
    setLoginOpen(true);
  };

  return (
    <ChromeProvider value={{ openLogin, closeLogin: () => setLoginOpen(false) }}>
      <div className="app-shell">
        <header className="topbar">
          <div className="topbar__inner">
            <div className="topbar__row topbar__row--primary">
              <div className="brand" aria-label="IWX">
                <span className="brand__line">IWX</span>
              </div>

              <nav className="topnav" aria-label="Market categories">
                <Link
                  className={`topnav__item${activePortfolio ? " topnav__item--active" : ""}`}
                  to="/portfolio"
                >
                  Portfolio
                </Link>
                <Link
                  className={`topnav__item${activeStations ? " topnav__item--active" : ""}`}
                  to="/"
                >
                  Stations
                </Link>
                <Link
                  className={`topnav__item${activeMarkets ? " topnav__item--active" : ""}`}
                  to="/markets"
                >
                  Markets
                </Link>
              </nav>

              <div className="topbar__actions">
                {session ? (
                  <>
                    <span className="topbar__session">User #{session.userId}</span>
                    <button className="topbar__login" type="button" onClick={logout}>
                      Log out
                    </button>
                  </>
                ) : (
                  <>
                    <button className="topbar__login" type="button" onClick={openLogin}>
                      Log in
                    </button>
                    <button className="topbar__signup" type="button" onClick={openSignup}>
                      Sign Up
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>
        </header>

        <main className="page">{children}</main>

        <LoginModal open={loginOpen} mode={authMode} onClose={() => setLoginOpen(false)} />
      </div>
    </ChromeProvider>
  );
}

export function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<StationsPage />} />
        <Route path="/stations" element={<Navigate to="/" replace />} />
        <Route path="/station/:stationId" element={<StationDetailPage />} />
        <Route path="/markets" element={<MarketsPage />} />
        <Route path="/contracts/:contractId" element={<ExchangePage />} />
        <Route path="/portfolio" element={<PortfolioPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  );
}
