import React, { useState, useEffect } from 'react';
import { BrowserRouter, Route, Switch, useParams } from "react-router-dom";
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import NavDropdown from 'react-bootstrap/NavDropdown';
import Container from 'react-bootstrap/Container';
import { Link } from 'react-router-dom';
import Overview from './Overview';
import './App.css';
import Details from './Details';
import 'bootstrap/dist/css/bootstrap.min.css';

function App() {
  const [error, setError] = useState(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [currentHost, setCurrentHost] = useState({});
  const [hosts, setHosts] = useState([]);

  let dataFetch = () => {
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
          if (result) {
            setHosts(result);
          } else {
            setHosts([]);
          }
        },
        (error) => {
          setError(error);
        }
      )
  }

  useEffect(() => {
    dataFetch();
    let interval = setInterval(() => {
      dataFetch();
    }, 10000);

    return () => {
      clearInterval(interval);
    }
  }, [])

  if (error) {
    return <pre>{JSON.stringify(error.message, null, 2)}</pre>;
  }

  if (!isLoaded) {
    return <div>Loading...</div>;
  }

  return (
    <BrowserRouter>
      <Container fluid>
        <Navbar bg="dark" expand="md" variant="dark">
          <Navbar.Toggle aria-controls="navbarSupportedContent" />
          <Navbar.Collapse id="navbarSupportedContent">
            <Nav className="mr-auto">
              <NavDropdown id="navbarDropdown" title={currentHost.hostname} active>
                {hosts.map((item, key) => {
                  return (
                    <NavDropdown.Item
                      key={key}
                      href={"http://" + item.publicIp}
                    >
                      {item.hostname}
                    </NavDropdown.Item>
                  );
                })}
              </NavDropdown>

              <Link className="nav-link" to="/">Home</Link>
            </Nav>
          </Navbar.Collapse>
        </Navbar>

        <Container fluid className="main-container">
          <Container fluid className="content-container">

            <Switch>
              <Route path="/hosts/:hostId">
                <HostDetails currentHost={currentHost} hosts={hosts} />
              </Route>

              <Route exact path="/">
                <Overview currentHost={currentHost} hosts={hosts} />
              </Route>

              <Route>
                <PageNotFound />
              </Route>
            </Switch>
          </Container>
        </Container>
      </Container>
    </BrowserRouter>
  );
}

function HostDetails() {
  let { hostId } = useParams();
  return <Details hostId={hostId} />;
}

function PageNotFound() {
  return <div className="text-center"><h3>Page Not Found</h3></div>;
}

export default App;
