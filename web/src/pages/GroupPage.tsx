import { useEffect, useState } from "react";
import { useParams } from "react-router";
import { api } from "../api/client";
import MemberList from "../components/MemberList";
import RecipientCard from "../components/RecipientCard";
import ChatPanel from "../components/ChatPanel";

interface GroupData {
  id: number;
  title: string;
  status: string;
  member_count?: number;
  members?: { name: string; is_me: boolean }[];
  is_organizer?: boolean;
  my_membership_id?: number;
}

export default function GroupPage({ userId }: { userId: number | null }) {
  const { inviteCode } = useParams<{ inviteCode: string }>();
  const [group, setGroup] = useState<GroupData | null>(null);
  const [recipient, setRecipient] = useState<{
    name: string;
    wishlist: string;
  } | null>(null);
  const [error, setError] = useState("");

  const groupId = group?.id ?? null;

  const loadGroup = async () => {
    try {
      const data = await api.getGroup(inviteCode!);
      setGroup(data);
    } catch {
      setError("Группа не найдена");
    }
  };

  useEffect(() => {
    loadGroup();
  }, [inviteCode]);

  useEffect(() => {
    if (group?.status === "drawn" && groupId && group.members) {
      api.getMyRecipient(groupId).then((data) => {
        setRecipient(data.recipient);
      });
    }
  }, [group?.status, groupId, group?.members]);

  const handleDraw = async () => {
    if (!groupId) return;
    try {
      await api.draw(groupId);
      loadGroup();
    } catch (e) {
      const err = e as { message?: string };
      setError(err.message || "Не удалось провести жеребьевку");
    }
  };

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-red-500">{error}</p>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Загрузка...</p>
      </div>
    );
  }

  if (!group.members) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <h1 className="text-2xl font-bold mb-2">{group.title}</h1>
          <p className="text-gray-600 mb-4">
            {group.member_count} участник(ов) |{" "}
            {group.status === "open"
              ? "Прием участников"
              : "Жеребьевка проведена"}
          </p>
          {group.status === "open" && userId && (
            <a
              href={`/g/${inviteCode}/join`}
              className="inline-block bg-red-600 text-white rounded-lg px-6 py-2 font-medium hover:bg-red-700"
            >
              Вступить
            </a>
          )}
          {!userId && (
            <p className="text-sm text-gray-500">
              Войдите, чтобы присоединиться
            </p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-4">
      <div className="max-w-2xl mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">{group.title}</h1>
          <span
            className={`text-sm px-3 py-1 rounded-full ${
              group.status === "open"
                ? "bg-yellow-100 text-yellow-800"
                : "bg-green-100 text-green-800"
            }`}
          >
            {group.status === "open"
              ? "Прием участников"
              : "Жеребьевка проведена"}
          </span>
        </div>

        {group.status === "open" && (
          <div className="bg-white rounded-lg shadow p-4">
            <p className="text-sm text-gray-600 mb-1">
              Ссылка для приглашения:
            </p>
            <code className="text-sm bg-gray-100 p-2 rounded block">
              {window.location.origin}/g/{inviteCode}
            </code>
          </div>
        )}

        <MemberList members={group.members} />

        {group.is_organizer && group.status === "open" && (
          <button
            onClick={handleDraw}
            className="w-full bg-green-600 text-white rounded-lg py-3 font-medium hover:bg-green-700"
          >
            Провести жеребьевку
          </button>
        )}

        {group.status === "drawn" && recipient && groupId && (
          <>
            <RecipientCard name={recipient.name} wishlist={recipient.wishlist} />
            <div className="space-y-4">
              <ChatPanel
                groupId={groupId}
                role="santa"
                title="Переписка с подопечным (ты — Санта)"
              />
              <ChatPanel
                groupId={groupId}
                role="recipient"
                title="Переписка с твоим Сантой"
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
