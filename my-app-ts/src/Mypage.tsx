import React, { useState, useEffect } from "react";
import { onAuthStateChanged, signOut } from "firebase/auth";
import { auth } from "./firebase";
import { useNavigate, Navigate } from "react-router-dom";
import { useParams } from "react-router-dom";
import { Link } from "react-router-dom";

type Channel = {
  id: string;
  name: string;
  description: string;
};

const Mypage: React.FC = () => {
  const [user, setUser] = useState<any>("");
  const [loading, setLoading] = useState(true);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [channelname, setChannelName] = useState("");
  const [channelDescription, setChannelDescription] = useState("");
  const [channelId, setChannelId] = useState("");
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
    navigate("/");
  };

  const fetchChannels = async () => {
    try {
      const res = await fetch(`http://localhost:8080/mypage?email=${email}`);
      if (!res.ok) {
        throw Error(`Failed to fetch users: ${res.status}`);
      }
      const channelsData: Channel[] = await res.json();
      setChannels(channelsData);
    } catch (err) {
      console.error(err + "1");
    }
  };

  useEffect(() => {
    fetchChannels();
  }, []);

  const makeChannel = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (channelname.length === 0) {
      alert("Please enter a channel name");
      return;
    }

    if (channelname.length > 50) {
      alert("Please enter a channel name shorter than 50 characters");
      return;
    }

    if (channelDescription.length === 0) {
      alert("Please enter a channel description");
      return;
    }

    try {
      const result = await fetch(`http://localhost:8080/makeChannel?email=${email}`, {
        method: "POST",
        body: JSON.stringify({
          name: channelname,
          description: channelDescription
        }),
      });
      if (!result.ok) {
        throw Error(`Failed to create user: ${result.status}`);
      }

      setChannelName("");
      setChannelDescription("");
      fetchChannels();
    } catch (err) {
      console.error(err + "2");
    }
  };

  const joinChannel = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (channelId.length !== 26) {
      alert("Please enter correct ID");
      return;
    }

    try {
      const result = await fetch(`http://localhost:8080/joinChannel?email=${email}`, {
        method: "POST",
        body: JSON.stringify({
          channel_id: channelId
        }),
      });
      if (!result.ok) {
        throw Error(`Failed to create user: ${result.status}`);
      }

      setChannelId("");
      fetchChannels();
    } catch (err) {
      console.error(err + "2");
    }
  };

  return (
    <>
      {!loading && (
        <>
          {!user ? (
            <Navigate to={`/`} />
          ) : (
            <>
              <h1>Channels</h1>
              <p>{user && user.email}</p>
        <ul>
            {channels.map((channel: Channel) => (
              <div key = {channel.id}>
              <div className="Map">
                <h4>
                  <Link to={`/channel/${channel.id}/${user.email}`}>Channel name: {channel.name}</Link>
                </h4>
                <h6>
                id: {channel.id}, description: {channel.description}
                </h6>
              </div>
              </div>
          ))}
        </ul>
        <form onSubmit={makeChannel}>
            <div>
              <label>New channel name</label>
              <input
                type="name"
                value={channelname}
                onChange={(e) => setChannelName(e.target.value)}
              />
            </div>
            <div>
              <label>Description</label>
              <input
                type="description"
                value={channelDescription}
                onChange={(e) => setChannelDescription(e.target.value)}
              />
            </div>
            <button type={"submit"}>
              Make new channel
            </button>
            </form>
            <form onSubmit={joinChannel}>
            <div>
              <label>Channel id</label>
              <input
                type="id"
                value={channelId}
                onChange={(e) => setChannelId(e.target.value)}
              />
            </div>
            <button type={"submit"}>
              Join channel
            </button>
            </form>
              <button onClick={logout}>ログアウト</button>
            </>
          )}
        </>
      )}
    </>
  );
};

export default Mypage;
