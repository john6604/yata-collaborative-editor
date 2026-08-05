const fallbackUsers = [];

function Sidebar({ isCollapsed, onToggle, users = fallbackUsers }) {
  return (
    <aside
      className={`flex shrink-0 flex-col border-l border-slate-200 bg-slate-950 text-white transition-all duration-200 ${
        isCollapsed ? 'w-16' : 'w-72'
      }`}
    >
      <div className="flex h-16 items-center justify-between border-b border-slate-800 px-4">
        {!isCollapsed && (
          <h2 className="text-sm font-semibold tracking-normal text-slate-100">Connected Users</h2>
        )}
        <button
          aria-expanded={!isCollapsed}
          aria-label={isCollapsed ? 'Expand connected users' : 'Collapse connected users'}
          className="ml-auto grid h-9 w-9 place-items-center rounded-md border border-slate-700 text-sm font-semibold text-slate-200 transition hover:bg-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-500"
          onClick={onToggle}
          type="button"
        >
          {isCollapsed ? '<' : '>'}
        </button>
      </div>

      {!isCollapsed && (
        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-5">
          {users.length > 0 ? (
            <ul className="space-y-3">
              {users.map((user) => (
                <li className="flex items-center gap-3" key={user.id}>
                  <span
                    aria-hidden="true"
                    className="h-9 w-9 shrink-0 rounded-full border border-white/20"
                    style={{ backgroundColor: user.color }}
                  />
                  <span className="min-w-0 truncate text-sm font-medium text-slate-100">
                    {user.name}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm leading-6 text-slate-400">No active users</p>
          )}
        </div>
      )}
    </aside>
  );
}

export default Sidebar;
