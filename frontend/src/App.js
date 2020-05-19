import React, {useState, useEffect} from 'react';
import {BrowserRouter, Route, Switch, useParams} from "react-router-dom";
import Layout from './Layout';
import './App.css';
import 'bootstrap/dist/css/bootstrap.min.css';

function App() {
  const [error, setError] = useState(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [currentHost, setCurrentHost] = useState({});
  const [hosts, setHosts] = useState([]);

  useEffect(() => {
    fetch("/api/health")
      .then(res => res.json())
      .then(
        (result) => {
          setIsLoaded(true);
          setCurrentHost(result);
        },
        (error) => {
          setIsLoaded(true);
          setError(error);
        }
      )
    fetch("/api/hosts")
      .then(res => res.json())
      .then(
        (result) => {
          setHosts(result);
        },
        (error) => {
          setError(error);
        }
      )
  }, [])

  if (error) {
    return <pre>{JSON.stringify(error.message, null, 2)}</pre>;
  }

  if (!isLoaded) {
    return <div>Loading...</div>;
  }

  return (
    <BrowserRouter>
      <Switch>
        <Route path="/hosts/:hostId">
          <HostDetails currentHost={currentHost} hosts={hosts} />
        </Route>

        <Route exact path="/">
          <Layout currentHost={currentHost} hosts={hosts} />
        </Route>

        <Route>
          <PageNotFound currentHost={currentHost} hosts={hosts} />
        </Route>
      </Switch>
    </BrowserRouter>
  );
}

function HostDetails(props) {
  let { hostId } = useParams();
  return <Layout currentHost={props.currentHost} hosts={props.hosts} hostId={hostId} />;
}

function PageNotFound(props) {
  return <Layout currentHost={props.currentHost} hosts={props.hosts} pageNotFound />;
}

export default App;
