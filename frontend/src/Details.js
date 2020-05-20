import React, { useState, useEffect } from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import ButtonGroup from 'react-bootstrap/ButtonGroup';
import Button from 'react-bootstrap/Button';
import { Line } from 'react-chartjs-2';
import moment from 'moment';
import 'bootstrap/dist/css/bootstrap.min.css';
import './Details.css';

function Details(props) {
    const [error, setError] = useState(null);
    const [isLoaded, setIsLoaded] = useState(false);
    const [host, setHost] = useState({});

    let dataFetch = () => {
        fetch("/api/hosts/" + props.hostId)
            .then(res => res.json())
            .then(
                (result) => {
                    setHost(result);
                    setIsLoaded(true);
                },
                (error) => {
                    setIsLoaded(true);
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
    }, [props.hostId])

    if (error) {
        return <pre>{JSON.stringify(error.message, null, 2)}</pre>;
    }

    if (!isLoaded) {
        return <div>Loading...</div>;
    }

    return (
        <Container fluid>
            <HostDetails host={host} />
            <br /><br />

            <Row>
                <Graph checksPublic={host.checks.public.http} checksInternal={host.checks.internal.http} type="http" />
                <Graph checksPublic={host.checks.public.icmp} checksInternal={host.checks.internal.icmp} type="icmp" />
                <Graph checksPublic={host.checks.public.tcp} checksInternal={host.checks.internal.tcp} type="tcp" />
                <Graph checksPublic={host.checks.public.udp} checksInternal={host.checks.internal.udp} type="udp" />
            </Row>
        </Container>
    );
}

function OverviewHostStatus(props) {
    if (props.status === 'success') {
        return <Button size="sm" variant="success" className="status-btn" disabled>{props.name}</Button>;
    }
    return <Button size="sm" variant="danger" className="status-btn" disabled>{props.name}</Button>;
}


function HostDetails(props) {
    return (
        <Row>
            <Container fluid>
                <Row>
                    <Col lg={12} className="text-center"><h3>{props.host.hostname}</h3></Col>
                </Row>
                <br />
                <Row>
                    <Col lg={6} className="text-center">
                        {props.host.publicIp}
                        <br />
                        <ButtonGroup>
                            <OverviewHostStatus name='HTTP' status={props.host.latestStatus.public.http} />
                            <OverviewHostStatus name='ICMP' status={props.host.latestStatus.public.icmp} />
                            <OverviewHostStatus name='TCP' status={props.host.latestStatus.public.tcp} />
                            <OverviewHostStatus name='UDP' status={props.host.latestStatus.public.udp} />
                        </ButtonGroup>
                    </Col>
                    <Col lg={6} className="text-center">
                        {props.host.internalIp}
                        <br />
                        <ButtonGroup>
                            <OverviewHostStatus name='HTTP' status={props.host.latestStatus.internal.http} />
                            <OverviewHostStatus name='ICMP' status={props.host.latestStatus.internal.icmp} />
                            <OverviewHostStatus name='TCP' status={props.host.latestStatus.internal.tcp} />
                            <OverviewHostStatus name='UDP' status={props.host.latestStatus.internal.udp} />
                        </ButtonGroup>
                    </Col>
                </Row>
            </Container>
        </Row>
    );
}

function Graph(props) {
    let publicData = [];
    let internalData = [];

    props.checksPublic.forEach(item => {
        publicData.push({
            t: moment.unix(parseInt(moment(item.checkedAt).format("X"))),
            y: item.status === "success" ? parseInt(item.responseTime / 1000 / 1000) : 0
        });
    });

    props.checksInternal.forEach(item => {
        internalData.push({
            t: moment.unix(parseInt(moment(item.checkedAt).format("X"))),
            y: item.status === "success" ? parseInt(item.responseTime / 1000 / 1000) : 0
        });
    });

    let options = {
        title: {
            display: true,
            fontSize: 24,
            fontColor: "#fff",
            text: props.type + " checks"
        },
        responsive: true,
        maintainAspectRatio: false,
        hoverMode: 'index',
        legend: {
            display: true
        },
        scales: {
            xAxes: [{
                type: 'time',
                time: {
                    unit: 'minute'
                }
            }],
            yAxes: [{
                display: true,
                ticks: {
                    suggestedMin: 0,
                    beginAtZero: true
                }
            }]
        }
    };

    let data = {
        datasets: [
            {
                label: "Public",
                data: publicData,
                borderColor: "#3e95cd",
                borderWidth: 2,
                backgroundColor: "rgba(132,99,255,0.05)",
                pointRadius: 1,
                lineTension: 0.2
            },
            {
                label: "Internal",
                data: internalData,
                borderColor: "#cc3e95",
                borderWidth: 2,
                backgroundColor: "rgba(255,99,132,0.05)",
                pointRadius: 1,
                lineTension: 0.2
            }
        ]
    };

    return (
        <div className="graph"><Line data={data} options={options} /></div>
    );
}

export default Details;