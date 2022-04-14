
#ifndef INSTRUMENT
#define INSTRUMENT

#include <map>
#include <string>
#include <fstream>
#include <mutex>
#include <ctime>
#include <random>


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
    string file;
    mutex m;
    
public:
    Instrument(){
       file = "stats_" + random_string() + ".out";  
    }

    void IncrementKey(string key){
       m.lock();
       keyToCounter[key]++;
       m.unlock();
    }

    void dumpStats(){
        m.lock();
        ofstream openedFile;
        openedFile.open(file, std::ios_base::app);
        openedFile << "StatsDump:" << std::time(0) << "\n";
        for(std::map<string,int>::iterator it = keyToCounter.begin(); it != keyToCounter.end(); it++){
           openedFile << it->first << "=" << it->second << "\n";
        }
        openedFile.close();
        m.unlock();
    }
};

#endif