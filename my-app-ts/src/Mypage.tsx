import React, { useState, useEffect } from "react";
import { onAuthStateChanged, signOut } from "firebase/auth";
import { auth } from "./firebase";
import { useNavigate, Navigate } from "react-router-dom";
import { useParams } from "react-router-dom";
import { BrowserRouter, Route, Routes, Link } from "react-router-dom";

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

const Mypage: React.FC = () => {
  const [user, setUser] = useState<any>("");
  const [loading, setLoading] = useState(true);
  const [messages, setMessages] = useState<Message[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const { email } = useParams();

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

  /*const fetchMessages = async () => {
    try {
      const res = await fetch(`http://localhost:8000/message`);
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
    fetchMessages();
  }, []);*/

  const fetchChannels = async () => {
    try {
      const res = await fetch(`http://localhost:8000/mypage?email=${email}`);
      if (!res.ok) {
        throw Error(`Failed to fetch users: ${res.status}`);
      }
      const channelsData: Channel[] = await res.json();
      setChannels(channelsData);
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
              {/*<ul>
            {messages.map((message: Message) => (
              <h6 key = {message.id}>
              {message.id}, {message.channel_id}, {message.user_id}, {message.content}
              </h6>
          ))}
        </ul>*/}
        <ul>
            {channels.map((channel: Channel) => (
              <h6 key = {channel.id}>
              {channel.id}, {channel.name}, {channel.description}
              <Link to={`/channel/${channel.id}/${user.email}`}>{channel.name}</Link>
              </h6>
          ))}
        </ul>
        
              <button onClick={logout}>ログアウト</button>
            </>
          )}
        </>
      )}
    </>
  );
};

export default Mypage;
