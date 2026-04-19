interface Member {
  name: string;
  is_me: boolean;
}

export default function MemberList({ members }: { members: Member[] }) {
  return (
    <div className="bg-white rounded-lg shadow p-4">
      <h2 className="font-bold text-lg mb-3">Участники ({members.length})</h2>
      <ul className="space-y-1">
        {members.map((m, i) => (
          <li key={i} className="flex items-center gap-2">
            <span className="text-gray-800">{m.name}</span>
            {m.is_me && (
              <span className="text-xs bg-red-100 text-red-700 px-2 py-0.5 rounded">
                это ты
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
