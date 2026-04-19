import { useState } from "react";
import { api } from "../api/client";

export default function LoginPage({
  onLogin: _onLogin,
}: {
  onLogin?: () => void;
}) {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await api.requestLink(email);
      setSent(true);
    } catch {
      setError("Не удалось отправить ссылку");
    }
  };

  if (sent) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <h1 className="text-2xl font-bold mb-4">Проверь почту</h1>
          <p className="text-gray-600">
            Мы отправили ссылку для входа на <strong>{email}</strong>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6 text-center">Тайный Санта</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            placeholder="you@example.com"
            required
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Получить ссылку для входа
          </button>
        </form>
      </div>
    </div>
  );
}
