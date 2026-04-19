import { useState, useEffect, useRef } from "react";
import { ChatSocket } from "../api/ws";
import { api } from "../api/client";

interface Message {
  id: number;
  from_me: boolean;
  body: string;
  created_at: string;
}

export default function ChatPanel({
  groupId,
  role,
  title,
}: {
  groupId: number;
  role: "santa" | "recipient";
  title: string;
}) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const socketRef = useRef<ChatSocket | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    api.getChatHistory(groupId, role).then((history) => {
      setMessages(history.reverse());
    });
  }, [groupId, role]);

  useEffect(() => {
    if (!socketRef.current) {
      socketRef.current = new ChatSocket(groupId, (msg) => {
        if (msg.type === "message" && msg.role === role) {
          setMessages((prev) => [
            ...prev,
            {
              id: msg.id!,
              from_me: msg.from_me!,
              body: msg.body!,
              created_at: msg.created_at!,
            },
          ]);
        }
      });
    }
    return () => {
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [groupId, role]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    socketRef.current?.send(role, input.trim());
    setInput("");
  };

  return (
    <div className="bg-white rounded-lg shadow flex flex-col h-80">
      <div className="px-4 py-2 border-b font-medium text-sm">{title}</div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex ${msg.from_me ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`rounded-lg px-3 py-2 max-w-xs ${
                msg.from_me
                  ? "bg-red-600 text-white"
                  : "bg-gray-100 text-gray-800"
              }`}
            >
              {msg.body}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <form onSubmit={handleSend} className="border-t p-2 flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-1"
          placeholder="Написать сообщение..."
          maxLength={2000}
        />
        <button
          type="submit"
          className="bg-red-600 text-white px-4 py-1 rounded-lg hover:bg-red-700"
        >
          Отправить
        </button>
      </form>
    </div>
  );
}
