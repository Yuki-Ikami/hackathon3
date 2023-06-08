import React, { useState, useEffect } from "react";
import { onAuthStateChanged, signOut } from "firebase/auth";
import { auth } from "./firebase";
import { useNavigate, Navigate } from "react-router-dom";
import { useParams } from "react-router-dom";

type Message = {
  id: string;
  channel_id: string;
  user_id: string;
  content: string;
};

type Channel = {
  id: string;
  name: string;
  description: string;
};

const Channel: React.FC = () => {
  const [user, setUser] = useState<any>("");
  const [loading, setLoading] = useState(true);
  const [messages, setMessages] = useState<Message[]>([]);
  const {channel_id, email} = useParams();
  const [postMessage, setPostMessage] = useState("");
  const [deleteId, setDeleteId] = useState("");
  const [editId, setEditId] = useState("");
  const [editedMessage, setEditedMessage] = useState("");

  const onPost = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!postMessage) {
      alert("Please enter a name shorter than 50 characters");
      return;
    }

    if (postMessage.length > 50) {
        alert("Please enter a name shorter than 50 characters");
        return;
      }

    try {
      const result = await fetch(`http://localhost:8080/channel?channelId=${channel_id}&email=${email}`, {
        method: "POST",
        body: JSON.stringify({
          content: postMessage
        }),
      });
      if (!result.ok) {
        throw Error(`Failed to create user: ${result.status}`);
      }

      setPostMessage("");
      fetchChannels();
    } catch (err) {
      console.error(err + "2");
    }
  };

  const onEdit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (editId.length != 26) {
      alert("Please enter correct ID");
      return;
    }

    try {
      const result = await fetch("http://localhost:8080/edit", {
        method: "EDIT",
        body: JSON.stringify({
          id: editId,
          message: editedMessage
        }),
      });
      if (!result.ok) {
        throw Error(`Failed to create user: ${result.status}`);
      }

      setEditId("");
      setEditedMessage("");
      fetchChannels();
    } catch (err) {
      console.error(err + "2");
    }
  };

  const onDelete = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (deleteId.length != 26) {
      alert("Please enter correct ID");
      return;
    }

    try {
      const result = await fetch("http://localhost:8080/delete", {
        method: "DELETE",
        body: JSON.stringify({
          id: deleteId,
        }),
      });
      if (!result.ok) {
        throw Error(`Failed to create user: ${result.status}`);
      }

      setDeleteId("");
      fetchChannels();
    } catch (err) {
      console.error(err + "2");
    }
  };

  useEffect(() => {
    onAuthStateChanged(auth, (currentUser) => {
      if (currentUser && currentUser.email) {
        setUser(currentUser);
      }
      setLoading(false);
    });
  }, []);

  const navigate = useNavigate();

  const logout = async () => {
    await signOut(auth);
    navigate("/login/");
  };

  const fetchChannels = async () => {
    try {
      const res = await fetch(`http://localhost:8080/message?channelId=${channel_id}?email=${email}`);
      if (!res.ok) {
        throw Error(`Failed to fetch users: ${res.status}`);
      }
      const messagesData: Message[] = await res.json();
      setMessages(messagesData);
    } catch (err) {
      console.error(err + "1");
    }
  };
//
  useEffect(() => {
    fetchChannels();
  }, []);

  return (
    <>
      {!loading && (
        <>
          {!user ? (
            <Navigate to={`/login/`} />
          ) : (
            <>
              <h1>マイページ</h1>
              <p>{user && user.email}</p>
              <ul>
              <h6>
              message_id, user_id, content
              </h6>
            {messages.map((message: Message) => (
              <h6 key = {message.id}>
              {message.id}, {message.user_id}, {message.content}
              </h6>
          ))}
        </ul>
        <form onSubmit={onPost}>
          <label>Message: </label>
          <input
            type={"message"}
            value={postMessage}
            onChange={(e) => setPostMessage(e.target.value)}
          />
          <button type={"submit"}>Post</button>
        </form>

        <form onSubmit={onEdit}>
          <label>Edit ID: </label>
          <input
            type={"id"}
            value={editId}
            onChange={(e) => setEditId(e.target.value)}
          />
          <label>Edit Message: </label>
          <input
            type={"message"}
            value={editedMessage}
            onChange={(e) => setEditedMessage(e.target.value)}
          />
          <button type={"submit"}>Edit</button>
        </form>

        <form onSubmit={onDelete}>
          <label>Delete ID: </label>
          <input
            type={"id"}
            value={deleteId}
            onChange={(e) => setDeleteId(e.target.value)}
          />
          <button type={"submit"}>Delete</button>
        </form>
              <button onClick={logout}>ログアウト</button>
            </>
          )}
        </>
      )}
    </>
  );
};

export default Channel;
