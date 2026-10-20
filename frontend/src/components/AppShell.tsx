import { type ReactNode } from 'react';
import { BookOpen, LogOut, History, UserCircle2, Menu, X } from 'lucide-react';
import { useSession } from '@/session';
import { Link } from '@/components/ui';
import type { Route } from '@/types';
import { useState } from 'react';

export function AppShell({ children }: { children: ReactNode }) {
  const { user, logout } = useSession();
  const [menuOpen, setMenuOpen] = useState(false);

  const navLinks: { to: Route; label: string }[] = [
    { to: { name: 'topics', params: {} }, label: 'Topics' },
  ];
  if (user) {
    navLinks.push({ to: { name: 'history', params: {} }, label: 'Quiz History' });
  }

  async function handleLogout() {
    await logout();
    setMenuOpen(false);
    window.location.hash = '#/topics';
  }

  return (
    <div className="min-h-screen flex flex-col bg-ink-50">
      <header className="sticky top-0 z-40 bg-white/85 backdrop-blur-md border-b border-ink-200/70">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
          <Link to={{ name: 'topics', params: {} }} className="flex items-center gap-2 group shrink-0">
            <div className="w-8 h-8 rounded-lg bg-ink-900 flex items-center justify-center text-white group-hover:bg-ink-800 transition-colors">
              <BookOpen className="w-4 h-4" />
            </div>
            <span className="font-semibold text-ink-900 tracking-tight hidden sm:block">Forge</span>
            <span className="text-[11px] text-ink-400 hidden md:block -ml-1 mt-0.5 font-mono">AI&nbsp;Engineering</span>
          </Link>

          <nav className="hidden md:flex items-center gap-1">
            {navLinks.map((l) => (
              <Link key={l.label} to={l.to} className="px-3 py-1.5 rounded-md text-sm text-ink-600 hover:text-ink-900 hover:bg-ink-100 transition-colors">
                {l.label}
              </Link>
            ))}
          </nav>

          <div className="hidden md:flex items-center gap-2">
            {user ? (
              <>
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-ink-50">
                  <UserCircle2 className="w-4 h-4 text-ink-400" />
                  <span className="text-sm text-ink-700 font-medium">{user.email}</span>
                  {user.isMaintainer && <span className="badge bg-brand-50 text-brand-700 text-[10px]">maintainer</span>}
                </div>
                <button onClick={handleLogout} className="btn-ghost text-sm">
                  <LogOut className="w-4 h-4" /> Log out
                </button>
              </>
            ) : (
              <>
                <Link to={{ name: 'login', params: {} }} className="btn-ghost text-sm">Log in</Link>
                <Link to={{ name: 'register', params: {} }} className="btn-primary text-sm">Sign up</Link>
              </>
            )}
          </div>

          <button
            className="md:hidden p-2 -mr-2 rounded-md text-ink-600 hover:bg-ink-100"
            onClick={() => setMenuOpen((v) => !v)}
            aria-label="Toggle menu"
            aria-expanded={menuOpen}
          >
            {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>

        {menuOpen && (
          <div className="md:hidden border-t border-ink-200/70 bg-white animate-fade-in">
            <div className="max-w-6xl mx-auto px-4 py-3 flex flex-col gap-1">
              {navLinks.map((l) => (
                <Link key={l.label} to={l.to} onClick={() => setMenuOpen(false)} className="px-3 py-2 rounded-md text-sm text-ink-700 hover:bg-ink-100">
                  {l.label}
                </Link>
              ))}
              <div className="h-px bg-ink-100 my-1" />
              {user ? (
                <>
                  <div className="px-3 py-2 text-sm text-ink-600 flex items-center gap-2">
                    <UserCircle2 className="w-4 h-4" /> {user.email}
                    {user.isMaintainer && <span className="badge bg-brand-50 text-brand-700 text-[10px]">maintainer</span>}
                  </div>
                  <button onClick={handleLogout} className="px-3 py-2 rounded-md text-sm text-left text-error-600 hover:bg-error-50 flex items-center gap-2">
                    <LogOut className="w-4 h-4" /> Log out
                  </button>
                </>
              ) : (
                <>
                  <Link to={{ name: 'login', params: {} }} onClick={() => setMenuOpen(false)} className="px-3 py-2 rounded-md text-sm text-ink-700 hover:bg-ink-100">Log in</Link>
                  <Link to={{ name: 'register', params: {} }} onClick={() => setMenuOpen(false)} className="px-3 py-2 rounded-md text-sm text-white bg-ink-900 hover:bg-ink-800">Sign up</Link>
                </>
              )}
            </div>
          </div>
        )}
      </header>

      <main className="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-8 sm:py-10">
        {children}
      </main>

      <footer className="border-t border-ink-200/70 bg-white">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-ink-400">
          <div className="flex items-center gap-2">
            <BookOpen className="w-3.5 h-3.5" />
            <span>Forge — AI Engineering Learning Platform</span>
          </div>
          <div className="flex items-center gap-4">
            <span>Prototype · simulated auth &amp; data</span>
            {user && (
              <Link to={{ name: 'history', params: {} }} className="hover:text-ink-600 flex items-center gap-1">
                <History className="w-3.5 h-3.5" /> History
              </Link>
            )}
          </div>
        </div>
      </footer>
    </div>
  );
}
