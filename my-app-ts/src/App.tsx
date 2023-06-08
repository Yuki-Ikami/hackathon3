import React from "react";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Register from "./Register";
import Login from "./Login";
import Mypage from "./Mypage";
import Channel from "./Channel";


function App() {
  return (
    <div className="container">
      <BrowserRouter>
        <Routes>
          <Route path="/register" element={<Register />} />
          <Route path="/login" element={<Login />} />
          <Route path="/:email" element={<Mypage />} />
          <Route path="/channel/:channel_id/:email" element={<Channel />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}

function Home() {
  return <h2>Home</h2>;
}

function About() {
  return <h2>About</h2>;
}

function Contact() {
  return <h2>Contact</h2>;
}

function NotFound() {
  return <h2>Not Found Page</h2>;
}

export default App;
