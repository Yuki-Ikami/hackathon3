import React, { useState, useEffect } from "react";
import { createUserWithEmailAndPassword, onAuthStateChanged } from "firebase/auth";
import { auth } from "./firebase";
import { Navigate, Link } from "react-router-dom";

const Register: React.FC = () => {
  const [registerEmail, setRegisterEmail] = useState("");
  const [registerPassword, setRegisterPassword] = useState("");
  const [registerName, setRegisterName] = useState("")

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!registerName) {
      alert("Please enter name");
      return;
    }

    if (registerName.length > 50) {
      alert("Please enter a name shorter than 50 characters");
      return;
    }

    try {
      await createUserWithEmailAndPassword(auth, registerEmail, registerPassword);
    } catch (error) {
      alert("正しく入力してください");
    }
    const fetchUsers = async () => {
      try {
        const res = await fetch("http://localhost:8000/user");
        if (!res.ok) {
          throw Error(`Failed to fetch users: ${res.status}`);
        }
        const users = await res.json();
        setUser(users);
      } catch (err) {
        console.error(err+"1");
      }
    };
  };

  const [user, setUser] = useState<any>("");

  useEffect(() => {
    onAuthStateChanged(auth, (currentUser) => {
      setUser(currentUser);
    });
  }, []);

  return (
    <>
      {user ? (
        <Navigate to={`/`} />
      ) : (
        <>
          <h1>新規登録</h1>
          <form onSubmit={handleSubmit}>
            <div>
              <label>メールアドレス</label>
              <input
                name="email"
                type="email"
                value={registerEmail}
                onChange={(e) => setRegisterEmail(e.target.value)}
              />
            </div>
            <div>
              <label>パスワード</label>
              <input
                name="password"
                type="password"
                value={registerPassword}
                onChange={(e) => setRegisterPassword(e.target.value)}
              />
            </div>
            <div>
              <label>名前</label>
              <input
                name="name"
                type="name"
                value={registerName}
                onChange={(e) => setRegisterName(e.target.value)}
              />
            </div>
            <button>登録する</button>
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
