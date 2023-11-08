import React, { useEffect, useState, useRef, useMemo } from 'react';
import { DashboardPartitionState } from './dashboard_state';
import rugbyPitchImage from '../../assets/pitch-background.png';
import ballImage from '../../assets/trywizard-ball.png';


const RugbyPitch: React.FC = () => {
  const [data, setData] = useState<{
    cumulativeTimesteps: number;
    partitionIndex: number;
    state: number[];
  }[]>([]);
  const [chartUpdatesEnabled, setChartUpdatesEnabled] = useState(true);
  const [selectedPartitionIndex, setSelectedPartitionIndex] = useState<string | null>(null);

  // create a memoized version of a function that generates the datasets
  const datasets = useMemo(() => {
    const result: { [partitionIndex: string]: {
      label: string;
      data: {
        time: number;
        state: number;
      }[];
    }[]} = {};

    data.forEach((datum) => {
      for (let index = 0; index < datum.state.length; index++) {
        const partitionIndexString = String(datum.partitionIndex);

        if (!(partitionIndexString in result)) {
          result[partitionIndexString] = [];
        }

        if (!(index in result[partitionIndexString])) {
          result[partitionIndexString].push({
            label: `Element ${index}`,
            data: [{
              time: datum.cumulativeTimesteps,
              state: datum.state[index],
            }],
          });
        } else {
          result[partitionIndexString][index].data.push({
            time: datum.cumulativeTimesteps,
            state: datum.state[index],
          })
        }
      }
    });

    return result;
  }, [data]);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:2112/dashboard');
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      console.log('Connected to WebSocket server');
    };

    ws.onmessage = async (event: MessageEvent) => {
      const decodedMessage = DashboardPartitionState.deserializeBinary(event.data);
      setData((prevData) => [
        ...prevData, {
          cumulativeTimesteps: decodedMessage.cumulative_timesteps, 
          partitionIndex: decodedMessage.partition_index, 
          state: decodedMessage.state
        },
      ]);
    };

    ws.onclose = () => {
      console.log('Disconnected from WebSocket server');
    };

    return () => {
      ws.close();
    };
  }, []);

  // Implement a function to update the selected partitionIndex
  const handlePartitionIndexChange = (partitionIndex: string) => {
    setSelectedPartitionIndex(partitionIndex);
  };

  return (
    <div>
      <div className="flex items-center justify-center h-64 border border-gray-300 rounded-lg p-4">
        <div className="rugby-pitch-container">
          <img src={rugbyPitchImage} alt="Rugby Pitch" className="rugby-pitch-image" />
          {data.map((datum, index) => (
            <img
              key={index}
              src={ballImage}
              alt="Rugby Ball"
              style={{
                position: 'absolute',
                top: `${datum.state[1]}px`,
                left: `${datum.state[0]}px`,
                transition: 'top 0.5s, left 0.5s', // CSS transition for smooth animation
              }}
              className={`ball-image ${String(datum.partitionIndex) === selectedPartitionIndex ? 'visible' : ''}`}
            />
          ))}
        </div>
      </div>
      <div>
        <div>
          {Object.entries(datasets).map(([k, v]) => (
            <button
              key={k}
              onClick={() => handlePartitionIndexChange(k)}
            >
              Show Partition {k}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
};

export default RugbyPitch;