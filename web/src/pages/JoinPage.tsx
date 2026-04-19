import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { api } from "../api/client";

export default function JoinPage() {
  const { inviteCode } = useParams<{ inviteCode: string }>();
  const [name, setName] = useState("");
  const [wishlist, setWishlist] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await api.joinGroup(inviteCode!, name, wishlist);
      navigate(`/g/${inviteCode}`);
    } catch {
      setError("Не удалось вступить в группу");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6">Присоединиться</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Твое имя
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            maxLength={50}
            required
          />
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Вишлист (что хочешь получить)
          </label>
          <textarea
            value={wishlist}
            onChange={(e) => setWishlist(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4 h-24"
            maxLength={2000}
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Вступить
          </button>
        </form>
      </div>
    </div>
  );
}
