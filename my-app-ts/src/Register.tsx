import React, { useState, useEffect } from "react";
import { createUserWithEmailAndPassword, onAuthStateChanged } from "firebase/auth";
import { auth } from "./firebase";
import { Navigate, Link } from "react-router-dom";

const Register: React.FC = () => {
  const [registerEmail, setRegisterEmail] = useState("");
  const [registerPassword, setRegisterPassword] = useState("");
  const [registerName, setRegisterName] = useState("");
  const [user, setUser] = useState<any>("");

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!registerName) {
      alert("Please enter a name");
      return;
    }

    if (registerName.length > 50) {
      alert("Please enter a name shorter than 50 characters");
      return;
    }

    try {
      await createUserWithEmailAndPassword(auth, registerEmail, registerPassword);
      await fetchUsers();
    } catch (error) {
      alert("Failed to register. Please check your input.");
    }
  };

  const fetchUsers = async () => {
    try {
      const res = await fetch("http://localhost:8000/register", {
        method: "POST",
        body: JSON.stringify({
          email: registerEmail,
          password: registerPassword,
          name: registerName
        }),
      });

      if (!res.ok) {
        throw Error(`Failed to fetch users: ${res.status}`);
      }

      const users = await res.json();
      setUser(users);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    onAuthStateChanged(auth, (currentUser) => {
      setUser(currentUser);
    });
  }, []);

  return (
    <>
      {user ? (
        <Navigate to={`/${user.email}`} />
      ) : (
        <>
          <h1>新規登録</h1>
          <form onSubmit={handleRegisterSubmit}>
            <div>
              <label htmlFor="email">メールアドレス</label>
              <input
                id="email"
                name="email"
                type="email"
                value={registerEmail}
                onChange={(e) => setRegisterEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="password">パスワード</label>
              <input
                id="password"
                name="password"
                type="password"
                value={registerPassword}
                autoComplete="new-password"
                onChange={(e) => setRegisterPassword(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="name">名前</label>
              <input
                id="name"
                name="name"
                type="name"
                value={registerName}
                onChange={(e) => setRegisterName(e.target.value)}
              />
            </div>
            <button type="button" onClick={handleRegisterSubmit}>
              登録する
            </button>
            <p>
              ログインは<Link to={`/login/`}>こちら</Link>
            </p>
          </form>
        </>
      )}
    </>
  );
};

export default Register;
