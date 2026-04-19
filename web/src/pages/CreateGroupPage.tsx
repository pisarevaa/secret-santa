import { useState } from "react";
import { useNavigate } from "react-router";
import { api } from "../api/client";

export default function CreateGroupPage() {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const group = await api.createGroup(title);
      navigate(`/g/${group.invite_code}`);
    } catch {
      setError("Не удалось создать группу");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6">Создать группу</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Название
          </label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            placeholder="Новый год 2026"
            maxLength={100}
            required
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Создать
          </button>
        </form>
      </div>
    </div>
  );
}
