
#ifndef INSTRUMENT
#define INSTRUMENT

#include <map>
#include <string>
#include <fstream>
#include <mutex>
#include <ctime>
#include <random>
#include <thread>
#include <memory>




using namespace std;

std::string random_string()
{
     std::string str("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz");

     std::random_device rd;
     std::mt19937 generator(rd());

     std::shuffle(str.begin(), str.end(), generator);

     return str.substr(0, 32);    // assumes 32 < number of characters in str         
}

class Instrument {
private:
    map<string, int> keyToCounter;
    map<string, int> previousCounterValue;
    string file;
    mutex m;
    unique_ptr<thread> statsThread;



    
public:
    Instrument(){
        file = "stats_" + random_string() + ".out";  
        dumpStats();
        Instrument* ptr = this;
        thread* tmp = new thread(&Instrument::writeStats, this);
        statsThread.reset(tmp);
    }

    ~Instrument(){
        
        // Doesnt matter if throws an error
        statsThread->~thread();

    }

    void writeStats(){
        while (true){
            std::this_thread::sleep_for(std::chrono::milliseconds(1000));
            dumpStats();
        }
    }

    void dumpStats(){
        m.lock();
        ofstream openedFile;
        openedFile.open(file, std::ios_base::app);
        openedFile << "StatsDump:" << std::time(0) << "\n";
        for(std::map<string,int>::iterator it = keyToCounter.begin(); it != keyToCounter.end(); it++){
           // openedFile << "Value:" << it->first << "=" << it->second << "\n";
           openedFile << "Rate Per Second: " << it->first << "=" << it->second - previousCounterValue[it->first] << "\n";
           previousCounterValue[it->first]  = it->second;
        }
        openedFile.close();
        m.unlock();
    }        

    void IncrementKey(string key){
       m.lock();
       keyToCounter[key]++;
       m.unlock();
    }
};






#endif


