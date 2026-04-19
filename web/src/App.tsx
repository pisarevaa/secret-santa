import { createBrowserRouter, RouterProvider } from "react-router";
import { useState, useEffect } from "react";
import { api } from "./api/client";
import LoginPage from "./pages/LoginPage";
import CreateGroupPage from "./pages/CreateGroupPage";
import GroupPage from "./pages/GroupPage";
import JoinPage from "./pages/JoinPage";

function App() {
  const [user, setUser] = useState<{
    user_id: number;
    email: string;
    name: string;
  } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getMe()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Загрузка...</p>
      </div>
    );
  }

  const router = createBrowserRouter([
    {
      path: "/",
      element: user ? (
        <CreateGroupPage />
      ) : (
        <LoginPage onLogin={() => window.location.reload()} />
      ),
    },
    {
      path: "/g/:inviteCode",
      element: <GroupPage userId={user?.user_id ?? null} />,
    },
    {
      path: "/g/:inviteCode/join",
      element: user ? (
        <JoinPage />
      ) : (
        <LoginPage onLogin={() => window.location.reload()} />
      ),
    },
  ]);

  return <RouterProvider router={router} />;
}

export default App;
